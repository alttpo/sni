package retroarch

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sni/cmd/sni/config"
	"sni/devices"
	"sni/devices/snes/mapping"
	"sni/protos/sni"
	"sni/udpclient"
	"sni/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

const hextable = "0123456789abcdef"

type readOperation struct {
	RequestAddress devices.AddressTuple
	DeviceAddress  devices.AddressTuple

	RequestSize  int
	ResponseData []byte
}

type writeOperation struct {
	RequestAddress devices.AddressTuple
	DeviceAddress  devices.AddressTuple

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
	stateLock sync.Mutex

	addr *net.UDPAddr

	readWriteTimeout time.Duration

	expectationLock  sync.Mutex
	outgoing         chan *rwRequest
	expectedIncoming chan *rwRequest

	version string
	useRCR  bool

	closeLock sync.Mutex
	closed    bool
}

func (c *RAClient) FatalError(cause error) devices.DeviceError {
	return devices.DeviceFatal(fmt.Sprintf("retroarch: %v", cause), cause)
}

func (c *RAClient) NonFatalError(cause error) devices.DeviceError {
	return devices.DeviceNonFatal(fmt.Sprintf("retroarch: %v", cause), cause)
}

// isCloseWorthy returns true if the error should close the connection
func isCloseWorthy(err error) bool {
	if errors.Is(err, net.ErrClosed) {
		return false
	}
	return devices.IsFatal(err)
}

func NewRAClient(addr *net.UDPAddr, name string, timeout time.Duration) *RAClient {
	c := &RAClient{
		addr:             addr,
		readWriteTimeout: timeout,
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
	c.closeLock.Lock()
	defer c.closeLock.Unlock()

	if !c.closed {
		err = c.UDPClient.Close()
		close(c.outgoing)
		close(c.expectedIncoming)
		c.closed = true
	}

	return
}

func (c *RAClient) Connect(addr *net.UDPAddr) (err error) {
	if err = c.UDPClient.Connect(addr); err != nil {
		return
	}

	return
}

func (c *RAClient) DetectLoopback(others []*RAClient) bool {
	for i := range others {
		other := others[i]

		// detect loopback condition:
		laddr := c.UDPClient.LocalAddr()
		raddr := other.UDPClient.RemoteAddr()
		if laddr == nil {
			continue
		}
		if raddr == nil {
			continue
		}
		if laddr.Port == raddr.Port {
			if laddr.IP.Equal(raddr.IP) {
				return true
			}
		}
	}

	return false
}

func (c *RAClient) GetId() string {
	return c.addr.String()
}

func (c *RAClient) Version() string {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	return c.version
}
func (c *RAClient) HasVersion() bool {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	return c.version != ""
}

func (c *RAClient) DetermineVersion() (err error) {
	var rsp []byte
	req := []byte("VERSION\n")
	if logDetector {
		log.Printf("retroarch: > %s", req)
	}
	rsp, err = c.WriteThenRead(req, time.Now().Add(c.readWriteTimeout))
	if err != nil {
		return
	}

	if rsp == nil {
		return fmt.Errorf("no response received")
	}

	if logDetector {
		log.Printf("retroarch: < %s", rsp)
	}

	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	c.version = strings.TrimSpace(string(rsp))

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

func (d *RAClient) GetStatus(ctx context.Context) (raStatus, coreName, romFileName string, romCRC32 uint32, err error) {
	// v1.9.2:
	//GET_STATUS
	//GET_STATUS CONTENTLESS
	//GET_STATUS
	//GET_STATUS PLAYING bsnes-mercury,o2-lttphack-emu-13.6.0,crc32=dae58be6
	//GET_STATUS
	//GET_STATUS PAUSED bsnes-mercury,o2-lttphack-emu-13.6.0,crc32=dae58be6

	// v1.9.0:
	//GET_STATUS
	//GET_STATUS PLAYING super_nes,o2-lttphack-emu-13.6.0,crc32=dae58be6

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(d.readWriteTimeout)
	}

	var rsp []byte
	req := []byte("GET_STATUS\n")
	if config.VerboseLogging {
		log.Printf("retroarch: > %s", req)
	}
	rsp, err = d.WriteThenRead(req, deadline)
	if err != nil {
		return
	}
	if config.VerboseLogging {
		log.Printf("retroarch: < %s", rsp)
	}

	// parse the response:
	var args string
	_, err = fmt.Fscanf(bytes.NewReader(rsp), "GET_STATUS %s %s", &raStatus, &args)
	if err != nil {
		return
	}

	// split the second arg by commas:
	argsArr := strings.Split(args, ",")
	if len(argsArr) >= 1 {
		coreName = argsArr[0]
	}
	if len(argsArr) >= 2 {
		romFileName = argsArr[1]
	}
	if len(argsArr) >= 3 {
		// e.g. "crc32=dae58be6"
		crc32 := argsArr[2]
		if strings.HasPrefix(crc32, "crc32=") {
			crc32 = crc32[len("crc32="):]

			var crc32_u64 uint64
			crc32_u64, err = strconv.ParseUint(crc32, 16, 32)
			if err == nil {
				romCRC32 = uint32(crc32_u64)
			}
		}
	}

	return
}

func (d *RAClient) FetchFields(ctx context.Context, fields ...sni.Field) (values []string, err error) {
	var raStatus string
	var coreName string
	var romFileName string
	var romCRC32 uint32

	raStatus, coreName, romFileName, romCRC32, err = d.GetStatus(ctx)
	if err != nil {
		return
	}

	for _, field := range fields {
		switch field {
		case sni.Field_DeviceName:
			values = append(values, "retroarch")
			break
		case sni.Field_DeviceVersion:
			values = append(values, d.version)
			break
		case sni.Field_DeviceStatus:
			values = append(values, raStatus)
			break
		case sni.Field_CoreName:
			values = append(values, coreName)
			break
		case sni.Field_RomFileName:
			values = append(values, romFileName)
			break
		case sni.Field_RomHashType:
			values = append(values, "crc32")
			break
		case sni.Field_RomHashValue:
			values = append(values, strconv.FormatUint(uint64(romCRC32), 16))
			break
		default:
			// unknown value; append empty string to maintain index association:
			values = append(values, "")
			break
		}
	}

	return
}

// RA 1.9.0 allows a maximum read size of 2723 bytes so we cut that off at 2048 to make division easier
const maxReadSize = 2048

func (c *RAClient) readCommand() string {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	if c.useRCR {
		return "READ_CORE_RAM"
	} else {
		return "READ_CORE_MEMORY"
	}
}

func (c *RAClient) writeCommand() string {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	if c.useRCR {
		return "WRITE_CORE_RAM"
	} else {
		return "WRITE_CORE_MEMORY"
	}
}

func (c *RAClient) RequiresMemoryMappingForAddressSpace(ctx context.Context, addressSpace sni.AddressSpace) (bool, error) {
	if addressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	if addressSpace == sni.AddressSpace_SnesABus {
		return false, nil
	}
	return true, nil
}

func (c *RAClient) RequiresMemoryMappingForAddress(ctx context.Context, address devices.AddressTuple) (bool, error) {
	if address.AddressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	if address.AddressSpace == sni.AddressSpace_SnesABus {
		return false, nil
	}
	return true, nil
}

func (c *RAClient) MultiReadMemory(ctx context.Context, reads ...devices.MemoryReadRequest) (mrsp []devices.MemoryReadResponse, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}
	outgoing := make([]*rwRequest, 0, len(reads)*2)

	// translate request addresses to device space:
	mrsp = make([]devices.MemoryReadResponse, len(reads))
	for j, read := range reads {
		mrsp[j] = devices.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			DeviceAddress: devices.AddressTuple{
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
					RequestAddress: devices.AddressTuple{
						Address:       read.RequestAddress.Address + offs,
						AddressSpace:  read.RequestAddress.AddressSpace,
						MemoryMapping: read.RequestAddress.MemoryMapping,
					},
					DeviceAddress: devices.AddressTuple{
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
					RequestAddress: devices.AddressTuple{
						Address:       read.RequestAddress.Address + offs,
						AddressSpace:  read.RequestAddress.AddressSpace,
						MemoryMapping: read.RequestAddress.MemoryMapping,
					},
					DeviceAddress: devices.AddressTuple{
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
	err = nil
	for _, rwreq := range outgoing {
		err = <-responses
		if err != nil {
			if derr, ok := err.(*readResponseError); ok {
				log.Printf("retroarch: read %#v returned error '%s'; filling response with $00\n", derr.Address, derr.Response)
				// fill response with 00 bytes:
				rwreq.Read.ResponseData = rwreq.Read.ResponseData[0:rwreq.Read.RequestSize]
				d := rwreq.Read.ResponseData
				for i := range d {
					d[i] = 0
				}
				err = nil
			}
			if err != nil {
				return
			}
		}

		mrsp[rwreq.index].Data = append(mrsp[rwreq.index].Data, rwreq.Read.ResponseData...)
	}

	return
}

func (c *RAClient) MultiWriteMemory(ctx context.Context, writes ...devices.MemoryWriteRequest) (mrsp []devices.MemoryWriteResponse, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	outgoing := make([]*rwRequest, 0, len(writes)*2)

	// translate addresses:
	mrsp = make([]devices.MemoryWriteResponse, len(writes))
	for j, write := range writes {
		mrsp[j] = devices.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: devices.AddressTuple{
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
					RequestAddress: devices.AddressTuple{
						Address:       write.RequestAddress.Address + offs,
						AddressSpace:  write.RequestAddress.AddressSpace,
						MemoryMapping: write.RequestAddress.MemoryMapping,
					},
					DeviceAddress: devices.AddressTuple{
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
					RequestAddress: devices.AddressTuple{
						Address:       write.RequestAddress.Address + offs,
						AddressSpace:  write.RequestAddress.AddressSpace,
						MemoryMapping: write.RequestAddress.MemoryMapping,
					},
					DeviceAddress: devices.AddressTuple{
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
	err = nil
	for _, rwreq := range outgoing {
		err = <-responses
		if err != nil {
			return
		}

		mrsp[rwreq.index].Size += rwreq.Write.ResponseSize
	}

	return
}

func (c *RAClient) handleOutgoing() {
	defer util.Recover()

	c.stateLock.Lock()
	useRCR := c.useRCR
	c.stateLock.Unlock()

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

			if config.VerboseLogging {
				log.Printf("retroarch: > %s", reqStr)
			}

			err := c.WriteWithDeadline([]byte(reqStr), rwreq.deadline)
			if err != nil {
				c.expectationLock.Unlock()
				rwreq.R <- err
				if isCloseWorthy(err) {
					_ = c.Close()
				}
				return
			}

			if useRCR && rwreq.isWrite {
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
	defer util.Recover()

	for rwreq := range c.expectedIncoming {
		rsp, err := c.ReadWithDeadline(rwreq.deadline)
		if err != nil {
			rwreq.R <- err
			if isCloseWorthy(err) {
				_ = c.Close()
			}
			break
		}

		if config.VerboseLogging {
			log.Printf("retroarch: < %s", rsp)
		}

		err = c.parseCommandResponse(rsp, rwreq)
		rwreq.R <- err
	}
}

type readResponseError struct {
	Address  uint32
	Response string
}

func (r *readResponseError) Error() string {
	return r.Response
}

func (r *readResponseError) IsFatal() bool {
	return false
}

func (c *RAClient) parseCommandResponse(rsp []byte, rwreq *rwRequest) (err error) {
	c.stateLock.Lock()
	useRCR := c.useRCR
	c.stateLock.Unlock()

	r := bytes.NewReader(rsp)

	var cmd string
	var addr uint32
	var n int
	n, err = fmt.Fscanf(r, "%s %x", &cmd, &addr)
	if n != 2 || cmd != rwreq.command || addr != rwreq.address {
		err = c.FatalError(fmt.Errorf("expected response starting with `%s %x` but got: `%s`", rwreq.command, rwreq.address, string(rsp)))
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
			if useRCR {
				err = &readResponseError{addr, ""}
				return
			} else {
				// READ_CORE_MEMORY returns an error description after -1
				// e.g. `READ_CORE_MEMORY 40ffb0 -1 no data for descriptor`
				var txt string
				txt, err = bufio.NewReader(t).ReadString('\n')
				if err != nil {
					log.Printf("could not read error text from %s response: %v; `%s`", cmd, err, string(rsp))
					err = c.FatalError(err)
					return
				}

				txt = strings.TrimSpace(txt)
				err = &readResponseError{addr, txt}
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
			err = c.FatalError(fmt.Errorf(
				"%s responded with unexpected length %d; expected %d; `%s`",
				cmd,
				wlen,
				len(rwreq.Write.RequestData),
				string(rsp),
			))
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
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	req := []byte("RESET\n")
	if config.VerboseLogging {
		log.Printf("retroarch: > %s", req)
	}
	err = c.WriteWithDeadline(req, deadline)
	return
}

func (c *RAClient) ResetToMenu(ctx context.Context) error {
	panic("implement me")
}

func (c *RAClient) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	return false, fmt.Errorf("capability unavailable")
}

func (c *RAClient) PauseToggle(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	req := []byte("PAUSE_TOGGLE\n")
	if config.VerboseLogging {
		log.Printf("retroarch: > %s", req)
	}
	err = c.WriteWithDeadline(req, deadline)
	return
}
