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
	rcrHasBusMapping  bool
	rcrTestsTried string

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
	defer c.stateLock.Unlock()
	c.stateLock.Lock()

	return c.version
}
func (c *RAClient) HasVersion() bool {
	defer c.stateLock.Unlock()
	c.stateLock.Lock()

	return c.version != ""
}

func (c *RAClient) DetermineVersionAndSystemAndApi() (systemId string, err error) {
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
		err = fmt.Errorf("no response received to VERSION request")
		return
	}

	if logDetector {
		log.Printf("retroarch: < %s", rsp)
	}

	defer c.stateLock.Unlock()
	c.stateLock.Lock()

	c.version = strings.TrimSpace(string(rsp))

	// parse the version string:
	var n int
	var major, minor, patch int
	n, err = fmt.Sscanf(c.version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil || n != 3 {
		return
	}
	err = nil

	var raStatus string
	raStatus, systemId, _, _, err = c.GetStatus(context.Background())
	if err != nil {
		if logDetector && raStatus == "CONTENTLESS" {
			// CONTENTLESS -> no way to present this as a specific system (super_nes, nintendo_64)
			log.Printf("retroarch: RA on port %d has no content loaded\n",
				c.UDPClient.RemoteAddr().Port)
		}
		return
	}

	// for RA version < 1.9.1, READ_CORE_RAM is the only API available. SNI has historically never
	// supported 'rcrHasBusMapping = false' behavior for these versions, which are now quite old.
	// setting rcrHasBusMapping = true means RA cores on RA < 1.9.1 that do not support bus mapping
	// will eventually error out on client memory requests
	if major < 1 {
		// 0.x.x
		c.useRCR = true
		c.rcrHasBusMapping = true
		return
	}
	if major == 1 && minor < 9 {
		// 1.0-8.x
		c.useRCR = true
		c.rcrHasBusMapping = true
		return
	}
	if major == 1 && minor == 9 && patch < 1 {
		// 1.9.0
		c.useRCR = true
		c.rcrHasBusMapping = true
		return
	}

	// for versions where 1.9.1 <= RA version < 1.10.1, the snes bus memory mapping is broken due to
	// a bug in RA. the ROM is essentially unreadable (gives back wrong data). so refuse to read it
	if (major == 1 && minor == 9) || (major == 1 && minor == 10 && patch < 1) {
		// 1.9.(1+), 1.10.0
		c.useRCR = true
		c.rcrHasBusMapping = false
	}

	// READ_CORE_MEMORY is implemented with valid data in RA >= 1.10.1. for those versions, use
	// this api if it demonstrates that it's really available
	// meanwhile, READ_CORE_RAM cannot provide bus addressing in these more recent versions
	c.rcrHasBusMapping = false
	if systemId == "super_nes" {
		// 1.10.1+
		err = c.DetermineSnesMemoryApiByTesting()
	}
	return
}

func (c *RAClient) DetermineSnesMemoryApiByTesting() (err error) {
	type TestCase struct {
		command         string
		ifSuccessEquals bool
		thenSetRcrTo    bool
		warnOnFail      bool
		shortName       string
	}
	testCases := []TestCase{
		// if $00:ffc0 is available, very likely ROM+SRAM+RAM are all available via RCM
		TestCase{command: "READ_CORE_MEMORY 00ffc0 32",
			ifSuccessEquals: true,
			thenSetRcrTo: false,
			warnOnFail: false,
			shortName: "RCM ROM"},
		// if RCR is responding, very likely SRAM+RAM are both available via RCR
		TestCase{command: "READ_CORE_RAM 00 32",
			ifSuccessEquals: true,
			thenSetRcrTo: true,
			warnOnFail: true,
			shortName: "RCR WRAM"},
		// RCR is disabled (it's tied to RA achievements somehow) and $00:ffc0 isn't available.
		// if we have RAM access, that's something.
		// (SRAM may be available too if a gRPC client knows its address and knows the game.)
		// if not, this is not a useful SNES device, and we fail
		TestCase{command: "READ_CORE_MEMORY 7e0000 32",
			ifSuccessEquals: true,
			thenSetRcrTo: false,
			warnOnFail: true,
			shortName: "RCM WRAM"},
	}

	matchFound := false
	c.rcrTestsTried = ""
	for _, testCase := range testCases {
		var testMemoryRsp []byte
		testMemoryReq := []byte(testCase.command + "\n")
		if logDetector {
			log.Printf("retroarch: > %s", testMemoryReq)
		}
		failed := false
		testMemoryRsp, err = c.WriteThenRead(testMemoryReq, time.Now().Add(c.readWriteTimeout))
		if err != nil {
			failed = true
		}
		if !failed {
			if testMemoryRsp == nil {
				failed = true
			}
		}
		if !failed {
			testMemoryRspString := strings.TrimSpace(string(testMemoryRsp))
			if logDetector {
				log.Printf("retroarch: < %s", testMemoryRspString)
			}
			if strings.Contains(testMemoryRspString, "-1") {
				failed = true
			}
		}
		if failed && testCase.warnOnFail {
			log.Printf("retroarch: Warning: snes test request '%s' failed. If connection " +
				"succeeds, things may still not work properly", testCase.command)
		}
		success := !failed
		c.rcrTestsTried += testCase.shortName + ", "
		if success == testCase.ifSuccessEquals {
			c.useRCR = testCase.thenSetRcrTo
			matchFound = true
			break
		}
	}

	if !matchFound {
		err = fmt.Errorf("all snes test read requests failed")
	}
	c.rcrTestsTried = strings.Trim(c.rcrTestsTried, ", ")

	return
}

func (c *RAClient) LogRCR() {
	log.Printf("retroarch: RA on port %d: useRCR=%t rcrTestsTried='%s'\n",
		c.UDPClient.RemoteAddr().Port, c.useRCR, c.rcrTestsTried)
}

func (d *RAClient) GetStatus(ctx context.Context) (raStatus, systemId, romFileName string, romCRC32 uint32, err error) {
	// v1.10.3:
	//GET_STATUS
	//GET_STATUS CONTENTLESS
	//GET_STATUS
	//GET_STATUS PLAYING super_nes,o2-lttphack-emu-13.6.0,crc32=dae58be6
	//GET_STATUS
	//GET_STATUS PAUSED super_nes,o2-lttphack-emu-13.6.0,crc32=dae58be6

	// v1.9.2 - different system_id:
	//GET_STATUS
	//GET_STATUS PLAYING bsnes-mercury,o2-lttphack-emu-13.6.0,crc32=dae58be6

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
		systemId = argsArr[0]
		if systemId != "super_nes" {
			// unknown system. but some RA versions around 1.9.2 put the core name as their system_id.
			// check for at least bsnes and snes9x
			if strings.Contains(strings.ToLower(systemId), "snes") {
				systemId = "super_nes"
			}
		}
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
	var romFileName string
	var romCRC32 uint32

	raStatus, _, romFileName, romCRC32, err = d.GetStatus(ctx)
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
	defer c.stateLock.Unlock()
	c.stateLock.Lock()

	if c.useRCR {
		return "READ_CORE_RAM"
	} else {
		return "READ_CORE_MEMORY"
	}
}

func (c *RAClient) writeCommand() string {
	defer c.stateLock.Unlock()
	c.stateLock.Lock()

	if c.useRCR {
		return "WRITE_CORE_RAM"
	} else {
		return "WRITE_CORE_MEMORY"
	}
}

func (c *RAClient) DefaultAddressSpace(context.Context) (sni.AddressSpace, error) {
	return defaultAddressSpace, nil
}

var ErrRAUnknownMapping = fmt.Errorf("only WRAM is available, and SRAM in FxPakPro-space requests, due to falling back to RA RCR (READ_CORE_RAM) with no bus mapping")

func (c *RAClient) RATranslateAddress(
	sourceAddress devices.AddressTuple,
	deviceSpace sni.AddressSpace,
) (deviceAddress uint32, err error) {

	if !c.useRCR || c.rcrHasBusMapping {
		// RA speaks bus mapping, so get the address in snes A-bus terms
		return mapping.TranslateAddress(
			sourceAddress,
			deviceSpace,
		)
	} else {
		// RA does not speak bus mapping, so translate RAM and ROM addresses to READ_CORE_RAM space:
		// 0-$1ffff: RAM
		// $20000-onward: SRAM
		switch sourceAddress.AddressSpace {
		case sni.AddressSpace_Raw:
			return sourceAddress.Address, nil
		case sni.AddressSpace_FxPakPro:
			// SRAM
			if sourceAddress.Address >= 0xE0_0000 && sourceAddress.Address <= 0xEF_FFFF {
				return (sourceAddress.Address - 0xE0_0000 + 0x2_0000), nil
			}
			// WRAM
			if sourceAddress.Address >= 0xF5_0000 && sourceAddress.Address <= 0xF6_FFFF {
				return (sourceAddress.Address - 0xF5_0000), nil
			}
			return 0, ErrRAUnknownMapping
		case sni.AddressSpace_SnesABus:
			// WRAM
			if sourceAddress.Address >= 0xF5_0000 && sourceAddress.Address <= 0xF6_FFFF {
				return (sourceAddress.Address - 0xF5_0000), nil
			}
			// there is no way to be sure if the request wanted SRAM:
			// - if it looks like a HiROM SRAM request, the client could have meant LoROM
			//   enhancement chip memory instead
			// - if it looks like a LoROM SRAM request, the client could have meant HiROM ROM space
			//   instead
			return 0, ErrRAUnknownMapping
		}
		return 0, ErrRAUnknownMapping
	}
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

		mrsp[j].DeviceAddress.Address, err = c.RATranslateAddress(
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

		mrsp[rwreq.index].Data = rwreq.Read.ResponseData
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

		mrsp[j].DeviceAddress.Address, err = c.RATranslateAddress(
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
