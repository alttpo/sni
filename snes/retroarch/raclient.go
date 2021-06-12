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

const readWriteTimeout = time.Second * 1

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
	rsp, err = c.WriteThenReadTimeout([]byte("VERSION\n"), readWriteTimeout)
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

func (c *RAClient) ReadMemory(context context.Context, read snes.MemoryReadRequest) (mrsp snes.MemoryReadResponse, err error) {
	busAddr := read.Address
	size := read.Size

	if size > maxReadSize {
		// TODO: make a batch request and concat the response pieces together:
		err = fmt.Errorf("read request size %d > max read size %d", size, maxReadSize)
		return
	}

	// TODO: detect -1 response
	var sb strings.Builder
	if c.useRCR {
		sb.WriteString("READ_CORE_RAM ")
	} else {
		sb.WriteString("READ_CORE_MEMORY ")
	}
	expectedAddr := busAddr
	sb.WriteString(fmt.Sprintf("%06x %d\n", expectedAddr, size))

	reqStr := sb.String()
	var rsp []byte

	rsp, err = c.WriteThenReadTimeout([]byte(reqStr), readWriteTimeout)
	if err != nil {
		return
	}

	r := bytes.NewReader(rsp)
	var data []byte
	data, err = c.parseReadMemoryResponse(r, expectedAddr, size)
	if err != nil {
		return
	}

	mrsp = snes.MemoryReadResponse{MemoryReadRequest: read, Data: data}
	return
}

func (c *RAClient) MultiReadMemory(context context.Context, reads ...snes.MemoryReadRequest) ([]snes.MemoryReadResponse, error) {
	panic("implement me")
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

		if c.useRCR {
			sb.WriteString("READ_CORE_RAM ")
		} else {
			sb.WriteString("READ_CORE_MEMORY ")
		}
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
	err = c.WriteTimeout([]byte(reqStr), readWriteTimeout)
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

		rsp, err = c.ReadTimeout(readWriteTimeout)
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

func (c *RAClient) WriteMemory(context context.Context, write snes.MemoryWriteRequest) error {
	panic("implement me")
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
		err = c.WriteTimeout([]byte(reqStr), readWriteTimeout)
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
			rsp, err = c.ReadTimeout(readWriteTimeout)
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
