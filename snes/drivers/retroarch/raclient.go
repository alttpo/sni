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
	"sync"
	"time"
)

const readWriteTimeout = time.Millisecond * 256

const hextable = "0123456789abcdef"

type readOperation struct {
	RequestAddress snes.AddressTuple
	DeviceAddress  snes.AddressTuple

	RequestSize  int
	ResponseData []byte
}

type writeOperation struct {
	RequestAddress snes.AddressTuple
	DeviceAddress  snes.AddressTuple

	RequestData  []byte
	ResponseSize int
}

type rwRequest struct {
	index    int
	deadline time.Time

	isWrite bool
	command string
	address uint32

	Read  readOperation
	Write writeOperation
	R     chan<- error
}

type RAClient struct {
	udpclient.UDPClient

	addr *net.UDPAddr

	expectationLock  sync.Mutex
	outgoing         chan *rwRequest
	expectedIncoming chan *rwRequest

	version string
	useRCR  bool
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
		addr:             addr,
		outgoing:         make(chan *rwRequest, 8),
		expectedIncoming: make(chan *rwRequest, 8),
	}
	udpclient.MakeUDPClient(name, &c.UDPClient)

	go c.handleIncoming()
	go c.handleOutgoing()

	return c
}

func (c *RAClient) IsClosed() bool { return c.UDPClient.IsClosed() }

func (c *RAClient) Close() (err error) {
	err = c.UDPClient.Close()
	close(c.outgoing)
	close(c.expectedIncoming)
	return
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

func (c *RAClient) readCommand() string {
	if c.useRCR {
		return "READ_CORE_RAM"
	} else {
		return "READ_CORE_MEMORY"
	}
}

func (c *RAClient) writeCommand() string {
	if c.useRCR {
		return "WRITE_CORE_RAM"
	} else {
		return "WRITE_CORE_MEMORY"
	}
}

func (c *RAClient) MultiReadMemory(ctx context.Context, reads ...snes.MemoryReadRequest) (mrsp []snes.MemoryReadResponse, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}
	outgoing := make([]*rwRequest, 0, len(reads)*2)

	// translate request addresses to device space:
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

		size := read.Size
		if size <= 0 {
			continue
		}
		rsp := mrsp[j]

		offs := uint32(0)
		for size > maxReadSize {
			// maxReadSize chunk:
			outgoing = append(outgoing, &rwRequest{
				index:    j,
				deadline: deadline,
				isWrite:  false,
				Read: readOperation{
					RequestAddress: snes.AddressTuple{
						Address:       read.RequestAddress.Address + offs,
						AddressSpace:  read.RequestAddress.AddressSpace,
						MemoryMapping: read.RequestAddress.MemoryMapping,
					},
					DeviceAddress: snes.AddressTuple{
						Address:       rsp.DeviceAddress.Address + offs,
						AddressSpace:  rsp.DeviceAddress.AddressSpace,
						MemoryMapping: rsp.DeviceAddress.MemoryMapping,
					},
					RequestSize:  maxReadSize,
					ResponseData: rsp.Data,
				},
			})
			offs += maxReadSize
			size -= maxReadSize
		}

		// remainder:
		if size > 0 {
			outgoing = append(outgoing, &rwRequest{
				index:    j,
				deadline: deadline,
				isWrite:  false,
				Read: readOperation{
					RequestAddress: snes.AddressTuple{
						Address:       read.RequestAddress.Address + offs,
						AddressSpace:  read.RequestAddress.AddressSpace,
						MemoryMapping: read.RequestAddress.MemoryMapping,
					},
					DeviceAddress: snes.AddressTuple{
						Address:       rsp.DeviceAddress.Address + offs,
						AddressSpace:  rsp.DeviceAddress.AddressSpace,
						MemoryMapping: rsp.DeviceAddress.MemoryMapping,
					},
					RequestSize:  size,
					ResponseData: rsp.Data,
				},
			})
		}
	}

	// make a channel to receive response errors:
	responses := make(chan error, len(outgoing))
	defer close(responses)

	// fire off all commands:
	for _, rwreq := range outgoing {
		rwreq.R = responses
		c.outgoing <- rwreq
	}

	// await all responses:
	for _, rwreq := range outgoing {
		err = <-responses
		if err != nil {
			if isCloseWorthy(err) {
				_ = c.Close()
			}
			return
		}

		mrsp[rwreq.index].Data = rwreq.Read.ResponseData
	}

	err = nil
	return
}

func (c *RAClient) MultiWriteMemory(ctx context.Context, writes ...snes.MemoryWriteRequest) (mrsp []snes.MemoryWriteResponse, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	outgoing := make([]*rwRequest, 0, len(writes)*2)

	// translate addresses:
	mrsp = make([]snes.MemoryWriteResponse, len(writes))
	for j, write := range writes {
		mrsp[j] = snes.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       0,
				AddressSpace:  sni.AddressSpace_SnesABus,
				MemoryMapping: write.RequestAddress.MemoryMapping,
			},
			Size: 0,
		}

		mrsp[j].DeviceAddress.Address, err = mapping.TranslateAddress(
			write.RequestAddress,
			sni.AddressSpace_SnesABus,
		)
		if err != nil {
			return nil, err
		}

		data := write.Data
		size := len(data)
		if size <= 0 {
			continue
		}
		rsp := &mrsp[j]

		offs := uint32(0)
		for size > maxReadSize {
			// maxReadSize chunk:
			outgoing = append(outgoing, &rwRequest{
				index:    j,
				deadline: deadline,
				isWrite:  true,
				Write: writeOperation{
					RequestAddress: snes.AddressTuple{
						Address:       write.RequestAddress.Address + offs,
						AddressSpace:  write.RequestAddress.AddressSpace,
						MemoryMapping: write.RequestAddress.MemoryMapping,
					},
					DeviceAddress: snes.AddressTuple{
						Address:       rsp.DeviceAddress.Address + offs,
						AddressSpace:  rsp.DeviceAddress.AddressSpace,
						MemoryMapping: rsp.DeviceAddress.MemoryMapping,
					},
					RequestData:  data[:maxReadSize],
					ResponseSize: maxReadSize,
				},
			})

			offs += maxReadSize
			size -= maxReadSize
			data = data[maxReadSize:]
		}

		// remainder:
		if size > 0 {
			outgoing = append(outgoing, &rwRequest{
				index:    j,
				deadline: deadline,
				isWrite:  true,
				Write: writeOperation{
					RequestAddress: snes.AddressTuple{
						Address:       write.RequestAddress.Address + offs,
						AddressSpace:  write.RequestAddress.AddressSpace,
						MemoryMapping: write.RequestAddress.MemoryMapping,
					},
					DeviceAddress: snes.AddressTuple{
						Address:       rsp.DeviceAddress.Address + offs,
						AddressSpace:  rsp.DeviceAddress.AddressSpace,
						MemoryMapping: rsp.DeviceAddress.MemoryMapping,
					},
					RequestData:  data[:size],
					ResponseSize: size,
				},
			})
		}
	}

	// make a channel to receive response errors:
	responses := make(chan error, len(outgoing))
	defer close(responses)

	// fire off all commands:
	for _, rwreq := range outgoing {
		rwreq.R = responses
		c.outgoing <- rwreq
	}

	// await all responses:
	for _, rwreq := range outgoing {
		err = <-responses
		if err != nil {
			if isCloseWorthy(err) {
				_ = c.Close()
			}
			return
		}

		mrsp[rwreq.index].Size += rwreq.Write.ResponseSize
	}

	return
}

func (c *RAClient) handleOutgoing() {
	for rwreq := range c.outgoing {
		var sb strings.Builder

		// build the proper command to send:
		if rwreq.isWrite {
			// write:
			rwreq.command = c.writeCommand()
			rwreq.address = rwreq.Write.DeviceAddress.Address
			_, _ = fmt.Fprintf(&sb, "%s %06x", rwreq.command, rwreq.address)

			// emit hex data to write:
			for _, v := range rwreq.Write.RequestData {
				sb.WriteByte(' ')
				sb.WriteByte(hextable[(v>>4)&0xF])
				sb.WriteByte(hextable[v&0xF])
			}
			sb.WriteByte('\n')
		} else {
			// read:
			rwreq.command = c.readCommand()
			rwreq.address = rwreq.Read.DeviceAddress.Address
			_, _ = fmt.Fprintf(&sb, "%s %06x %d\n", rwreq.command, rwreq.address, rwreq.Read.RequestSize)
		}

		// lock around sending the command and sending the expectation to read its response:
		c.expectationLock.Lock()
		{
			reqStr := sb.String()
			//log.Printf("retroarch: > %s", reqStr)

			err := c.WriteWithDeadline([]byte(reqStr), rwreq.deadline)
			if err != nil {
				c.expectationLock.Unlock()
				rwreq.R <- err
				return
			}

			if c.useRCR && rwreq.isWrite {
				// fake a response since we don't get any from WRITE_CORE_RAM:
				rwreq.R <- nil
			} else {
				// we're now expecting an incoming response:
				c.expectedIncoming <- rwreq
			}
		}
		c.expectationLock.Unlock()
	}
}

func (c *RAClient) handleIncoming() {
	for rwreq := range c.expectedIncoming {
		rsp, err := c.ReadWithDeadline(rwreq.deadline)
		if err != nil {
			rwreq.R <- err
			continue
		}

		//log.Printf("retroarch: < %s\n", string(rsp))

		err = c.parseCommandResponse(rsp, rwreq)
		rwreq.R <- err
	}
}

func (c *RAClient) parseCommandResponse(rsp []byte, rwreq *rwRequest) (err error) {
	r := bytes.NewReader(rsp)

	var cmd string
	var addr uint32
	var n int
	n, err = fmt.Fscanf(r, "%s %x", &cmd, &addr)
	if n != 2 || cmd != rwreq.command || addr != rwreq.address {
		err = fmt.Errorf("expected response starting with `%s %x` but got: `%s`", rwreq.command, rwreq.address, string(rsp))
		return
	}
	err = nil

	// handle `-1` and subsequent error text:
	{
		t := bytes.NewReader(rsp[r.Size()-int64(r.Len()):])
		var test int
		n, err = fmt.Fscanf(t, "%d", &test)
		if n == 1 && test < 0 {
			// read a -1:
			if c.useRCR {
				err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf("%s responded with error", cmd))
				return
			} else {
				// READ_CORE_MEMORY returns an error description after -1
				// e.g. `READ_CORE_MEMORY 40ffb0 -1 no data for descriptor`
				var txt string
				txt, err = bufio.NewReader(t).ReadString('\n')
				if err != nil {
					log.Printf("could not read error text from %s response: %v; `%s`", cmd, err, string(rsp))
					err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf("retroarch: unknown error"))
					return
				}

				txt = strings.TrimSpace(txt)
				err = fmt.Errorf("%s error '%s'", cmd, txt)
				if txt == "no data for descriptor" {
					err = snes.WithCode(codes.InvalidArgument, err)
				} else {
					// e.g. "no memory map defined" means no ROM loaded or emulator core doesn't support READ_CORE_MEMORY
					err = snes.WithCode(codes.FailedPrecondition, err)
				}
				return
			}
		}
		// not a -1:
		err = nil
	}

	if rwreq.isWrite {
		// write:
		var wlen int
		n, err = fmt.Fscanf(r, " %v\n", &wlen)
		if n != 1 {
			return
		}

		if wlen != len(rwreq.Write.RequestData) {
			err = fmt.Errorf(
				"%s responded with unexpected length %d; expected %d; `%s`",
				cmd,
				wlen,
				len(rwreq.Write.RequestData),
				string(rsp),
			)
			return
		}

		rwreq.Write.ResponseSize = wlen
		err = nil
		return
	} else {
		// read:
		data := make([]byte, 0, maxReadSize)
		for {
			var v byte
			n, err = fmt.Fscanf(r, " %02x", &v)
			if err != nil || n == 0 {
				break
			}
			data = append(data, v)
		}

		rwreq.Read.ResponseData = append(rwreq.Read.ResponseData, data...)

		err = nil
		return
	}
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