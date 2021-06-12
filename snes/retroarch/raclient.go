package retroarch

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"sni/snes"
	"sni/snes/lorom"
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
	// build multiple requests:
	var sb strings.Builder
	for _, read := range reads {
		size := read.Size
		if size <= 0 {
			continue
		}

		// TODO: support multiple ROM mappings
		addr := lorom.PakAddressToBus(read.Address)
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
		return
	}

	// responses come in multiple packets:
	mrsp = make([]snes.MemoryReadResponse, 0, len(reads))
	for _, read := range reads {
		size := read.Size
		if size <= 0 {
			continue
		}

		rrsp := snes.MemoryReadResponse{
			MemoryReadRequest: read,
			Data:              make([]byte, 0, read.Size),
		}

		// read chunks until complete:
		addr := lorom.PakAddressToBus(read.Address)
		for size > 0 {
			// parse ASCII response:
			rsp, err = c.ReadWithDeadline(deadline)
			if err != nil {
				return
			}
			var data []byte
			data, err = c.parseReadMemoryResponse(bytes.NewReader(rsp), addr, maxReadSize)
			if err != nil {
				return
			}

			// append response data:
			rrsp.Data = append(rrsp.Data, data...)

			addr += uint32(len(data))
			size -= len(data)
		}

		mrsp = append(mrsp, rrsp)
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

func (c *RAClient) ReadMemoryBatch(batch []snes.Read, keepAlive snes.KeepAlive) (err error) {
	// build multiple requests:
	var sb strings.Builder
	for _, req := range batch {
		// nowhere to put the response?
		completed := req.Completion
		if completed == nil {
			continue
		}

		_, _ = c.readCommand(&sb)
		expectedAddr := lorom.PakAddressToBus(req.Address)
		sb.WriteString(fmt.Sprintf("%06x %d\n", expectedAddr, req.Size))
	}

	reqStr := sb.String()
	var rsp []byte

	defer func() {
		c.Unlock()
	}()
	c.Lock()

	// send all commands up front in one packet:
	err = c.WriteWithDeadline([]byte(reqStr), time.Now().Add(readWriteTimeout))
	if err != nil {
		return
	}
	if keepAlive != nil {
		keepAlive <- struct{}{}
	}

	// responses come in multiple packets:
	for _, req := range batch {
		// nowhere to put the response?
		completed := req.Completion
		if completed == nil {
			continue
		}

		rsp, err = c.ReadWithDeadline(time.Now().Add(readWriteTimeout))
		if err != nil {
			return
		}
		if keepAlive != nil {
			keepAlive <- struct{}{}
		}

		expectedAddr := lorom.PakAddressToBus(req.Address)

		// parse ASCII response:
		r := bytes.NewReader(rsp)
		var data []byte
		data, err = c.parseReadMemoryResponse(r, expectedAddr, int(req.Size))
		if err != nil {
			continue
		}

		completed(snes.Response{
			IsWrite: false,
			Address: req.Address,
			Size:    req.Size,
			Extra:   req.Extra,
			Data:    data,
		})
	}

	err = nil
	return
}

func (c *RAClient) parseReadMemoryResponse(r *bytes.Reader, expectedAddr uint32, size int) (data []byte, err error) {
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

func (c *RAClient) MultiWriteMemory(context context.Context, writes ...snes.MemoryWriteRequest) error {
	panic("implement me")
}

func (c *RAClient) WriteMemoryBatch(batch []snes.Write, keepAlive snes.KeepAlive) (err error) {
	for _, req := range batch {
		var sb strings.Builder

		if c.useRCR {
			sb.WriteString("WRITE_CORE_RAM ")
		} else {
			sb.WriteString("WRITE_CORE_MEMORY ")
		}
		writeAddress := lorom.PakAddressToBus(req.Address)
		sb.WriteString(fmt.Sprintf("%06x ", writeAddress))
		// emit hex data:
		lasti := len(req.Data) - 1
		for i, v := range req.Data {
			sb.WriteByte(hextable[(v>>4)&0xF])
			sb.WriteByte(hextable[v&0xF])
			if i < lasti {
				sb.WriteByte(' ')
			}
		}
		sb.WriteByte('\n')
		reqStr := sb.String()

		log.Printf("retroarch: > %s", reqStr)
		err = c.WriteWithDeadline([]byte(reqStr), time.Now().Add(readWriteTimeout))
		if err != nil {
			return
		}
		if keepAlive != nil {
			keepAlive <- struct{}{}
		}
	}

	if !c.useRCR {
		for _, req := range batch {
			writeAddress := lorom.PakAddressToBus(req.Address)

			// expect a response from WRITE_CORE_MEMORY
			var rsp []byte
			rsp, err = c.ReadWithDeadline(time.Now().Add(readWriteTimeout))
			if err != nil {
				return
			}
			log.Printf("retroarch: < %s", rsp)
			if keepAlive != nil {
				keepAlive <- struct{}{}
			}

			var addr uint32
			var wlen int
			var n int
			r := bytes.NewReader(rsp)
			n, err = fmt.Fscanf(r, "WRITE_CORE_MEMORY %x %v\n", &addr, &wlen)
			if n != 2 {
				return
			}
			if addr != writeAddress {
				err = fmt.Errorf("retroarch: write_core_memory returned unexpected address %06x; expected %06x", addr, writeAddress)
				return
			}
			if wlen != len(req.Data) {
				err = fmt.Errorf("retroarch: write_core_memory returned unexpected length %d; expected %d", wlen, len(req.Data))
				return
			}

			completed := req.Completion
			if completed != nil {
				completed(snes.Response{
					IsWrite: true,
					Address: req.Address,
					Size:    req.Size,
					Extra:   req.Extra,
					Data:    req.Data,
				})
			}
		}
	}

	return
}
