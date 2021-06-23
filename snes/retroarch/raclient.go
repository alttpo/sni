package retroarch

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"log"
	"net"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"sni/udpclient"
	"strings"
	"time"
)

const readWriteTimeout = time.Millisecond * 256

const hextable = "0123456789abcdef"

type RAClient struct {
	udpclient.UDPClient

	addr *net.UDPAddr

	version string
	useRCR  bool
	mapping sni.MemoryMapping
}

// isCloseWorthy returns true if the error should close the connection
func isCloseWorthy(err error) bool {
	var coded *snes.CodedError
	if errors.As(err, &coded) {
		if coded.Code == codes.Internal {
			return true
		}
		return false
	}
	if errors.Is(err, net.ErrClosed) {
		return false
	}
	return true
}

func NewRAClient(addr *net.UDPAddr, name string) *RAClient {
	c := &RAClient{
		addr: addr,
	}
	udpclient.MakeUDPClient(name, &c.UDPClient)
	return c
}

func (c *RAClient) GetId() string {
	return c.addr.String()
}

func (c *RAClient) Version() string  { return c.version }
func (c *RAClient) HasVersion() bool { return c.version != "" }

func (c *RAClient) DetermineVersion() (err error) {
	var rsp []byte
	rsp, err = c.WriteThenRead([]byte("VERSION\n"), time.Now().Add(readWriteTimeout))
	if err != nil {
		return
	}

	if rsp == nil {
		return
	}

	log.Printf("retroarch: version %s", string(rsp))
	c.version = string(rsp)

	// parse the version string:
	var n int
	var major, minor, patch int
	n, err = fmt.Sscanf(c.version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil || n != 3 {
		return
	}
	err = nil

	// use READ_CORE_RAM for <= 1.9.0, use READ_CORE_MEMORY otherwise:
	c.useRCR = false
	if major < 1 {
		// 0.x.x
		c.useRCR = true
		return
	} else if major > 1 {
		// 2+.x.x
		c.useRCR = false
		return
	}
	if minor < 9 {
		// 1.0-8.x
		c.useRCR = true
		return
	} else if minor > 9 {
		// 1.10+.x
		c.useRCR = false
		return
	}
	if patch < 1 {
		// 1.9.0
		c.useRCR = true
		return
	}

	// 1.9.1+
	return
}

// RA 1.9.0 allows a maximum read size of 2723 bytes so we cut that off at 2048 to make division easier
const maxReadSize = 2048

func (c *RAClient) MultiReadMemory(context context.Context, reads ...snes.MemoryReadRequest) (mrsp []snes.MemoryReadResponse, err error) {
	// translate addresses:
	mrsp = make([]snes.MemoryReadResponse, len(reads))
	for j, read := range reads {
		mrsp[j] = snes.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       0,
				AddressSpace:  sni.AddressSpace_SnesABus,
				MemoryMapping: read.RequestAddress.MemoryMapping,
			},
			Data: make([]byte, 0, read.Size),
		}

		mrsp[j].DeviceAddress.Address, err = mapping.TranslateAddress(
			read.RequestAddress,
			sni.AddressSpace_SnesABus,
		)
		if err != nil {
			return nil, err
		}
	}

	// build multiple requests:
	var sb strings.Builder
	for j, read := range reads {
		size := read.Size
		if size <= 0 {
			continue
		}

		addr := mrsp[j].DeviceAddress.Address
		for size > maxReadSize {
			_, _ = c.readCommand(&sb)
			sb.WriteString(fmt.Sprintf("%06x %d\n", addr, maxReadSize))
			addr += maxReadSize
			size -= maxReadSize
		}
		if size > 0 {
			_, _ = c.readCommand(&sb)
			sb.WriteString(fmt.Sprintf("%06x %d\n", addr, size))
		}
	}

	reqStr := sb.String()
	var rsp []byte

	deadline, ok := context.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	// send all commands up front in one packet:
	err = c.WriteWithDeadline([]byte(reqStr), deadline)
	if err != nil {
		if isCloseWorthy(err) {
			_ = c.Close()
		}
		return
	}

	// responses come in multiple packets:
	for j, read := range reads {
		size := read.Size
		if size <= 0 {
			continue
		}

		rrsp := &mrsp[j]

		// read chunks until complete:
		addr := mrsp[j].DeviceAddress.Address
		for size > 0 {
			// parse ASCII response:
			rsp, err = c.ReadWithDeadline(deadline)
			if err != nil {
				if isCloseWorthy(err) {
					_ = c.Close()
				}
				return
			}
			var data []byte
			data, err = c.parseReadMemoryResponse(rsp, addr, maxReadSize)
			if err != nil {
				if isCloseWorthy(err) {
					_ = c.Close()
				}
				return
			}

			// append response data:
			rrsp.Data = append(rrsp.Data, data...)

			addr += uint32(len(data))
			size -= len(data)
		}
	}

	err = nil
	return
}

func (c *RAClient) readCommand(sb *strings.Builder) (int, error) {
	if c.useRCR {
		return sb.WriteString("READ_CORE_RAM ")
	} else {
		return sb.WriteString("READ_CORE_MEMORY ")
	}
}

func (c *RAClient) writeCommand(sb *strings.Builder) (int, error) {
	if c.useRCR {
		return sb.WriteString("WRITE_CORE_RAM ")
	} else {
		return sb.WriteString("WRITE_CORE_MEMORY ")
	}
}

func (c *RAClient) parseReadMemoryResponse(rsp []byte, expectedAddr uint32, size int) (data []byte, err error) {
	r := bytes.NewReader(rsp)

	var n int
	var addr uint32
	if c.useRCR {
		n, err = fmt.Fscanf(r, "READ_CORE_RAM %x", &addr)
	} else {
		n, err = fmt.Fscanf(r, "READ_CORE_MEMORY %x", &addr)
	}
	if err != nil {
		return
	}

	{
		t := bytes.NewReader(rsp[r.Size()-int64(r.Len()):])
		var test int
		n, err = fmt.Fscanf(t, "%d", &test)
		if n == 1 && test < 0 {
			// read a -1:
			if c.useRCR {
				err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf("retroarch: unknown error"))
				return
			} else {
				// READ_CORE_MEMORY returns an error description after -1
				// e.g. `READ_CORE_MEMORY 40ffb0 -1 no data for descriptor`
				var txt string
				txt, err = bufio.NewReader(t).ReadString('\n')
				if err != nil {
					log.Printf("retroarch: error attempting to read RA error text from response: %v; `%s`", err, string(rsp))
					err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf("retroarch: unknown error"))
					return
				}

				txt = strings.TrimSpace(txt)
				err = fmt.Errorf("retroarch: READ_CORE_MEMORY error '%s'", txt)
				if txt == "no data for descriptor" {
					err = snes.WithCode(codes.InvalidArgument, err)
				} else {
					// e.g. "no memory map defined" means no ROM loaded or emulator core doesn't support READ_CORE_MEMORY
					err = snes.WithCode(codes.FailedPrecondition, err)
				}
				return
			}
		}
		err = nil
	}

	if addr != expectedAddr {
		err = fmt.Errorf("retroarch: read response for wrong request %06x != %06x", addr, expectedAddr)
		return
	}

	data = make([]byte, 0, size)
	for {
		var v byte
		n, err = fmt.Fscanf(r, " %02x", &v)
		if err != nil || n == 0 {
			break
		}
		data = append(data, v)
	}

	err = nil
	return
}

func (c *RAClient) MultiWriteMemory(context context.Context, writes ...snes.MemoryWriteRequest) (mrsps []snes.MemoryWriteResponse, err error) {
	deadline, ok := context.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	// translate addresses:
	mrsps = make([]snes.MemoryWriteResponse, len(writes))
	for i, write := range writes {
		mrsps[i] = snes.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       0,
				AddressSpace:  sni.AddressSpace_SnesABus,
				MemoryMapping: write.RequestAddress.MemoryMapping,
			},
			Size: len(write.Data),
		}

		mrsps[i].DeviceAddress.Address, err = mapping.TranslateAddress(
			write.RequestAddress,
			sni.AddressSpace_SnesABus,
		)
		if err != nil {
			return nil, err
		}
	}

	for j, write := range writes {
		var sb strings.Builder

		_, _ = c.writeCommand(&sb)
		sb.WriteString(fmt.Sprintf("%06x", mrsps[j].DeviceAddress.Address))

		// emit hex data:
		for _, v := range write.Data {
			sb.WriteByte(' ')
			sb.WriteByte(hextable[(v>>4)&0xF])
			sb.WriteByte(hextable[v&0xF])
		}
		sb.WriteByte('\n')
		reqStr := sb.String()

		//log.Printf("retroarch: > %s", reqStr)
		err = c.WriteWithDeadline([]byte(reqStr), deadline)
		if err != nil {
			if isCloseWorthy(err) {
				_ = c.Close()
			}
			return
		}
	}

	if c.useRCR {
		// don't read any responses for READ_CORE_RAM
		return
	}

	for j, write := range writes {
		// expect a response from WRITE_CORE_MEMORY
		var rsp []byte
		rsp, err = c.ReadWithDeadline(deadline)
		if err != nil {
			_ = c.Close()
			return
		}
		//log.Printf("retroarch: < %s", rsp)

		writeAddress := mrsps[j].DeviceAddress.Address

		var addr uint32
		var wlen int
		var n int
		r := bytes.NewReader(rsp)
		n, err = fmt.Fscanf(r, "WRITE_CORE_MEMORY %x %v\n", &addr, &wlen)
		if n != 2 {
			return
		}
		if addr != writeAddress {
			err = fmt.Errorf("retroarch: write_core_memory returned unexpected address %06x; expected %06x; response:\n%s", addr, writeAddress, string(rsp))
			_ = c.Close()
			return
		}
		if wlen != len(write.Data) {
			err = fmt.Errorf("retroarch: write_core_memory returned unexpected length %d; expected %d; response:\n%s", wlen, len(write.Data), string(rsp))
			_ = c.Close()
			return
		}
	}

	return
}

func (c *RAClient) ResetSystem(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	err = c.WriteWithDeadline([]byte("RESET\n"), deadline)
	return
}

func (c *RAClient) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	return false, fmt.Errorf("capability unavailable")
}

func (c *RAClient) PauseToggle(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	err = c.WriteWithDeadline([]byte("PAUSE_TOGGLE\n"), deadline)
	return
}
