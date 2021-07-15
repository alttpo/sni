package emunw

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"strings"
	"sync"
	"time"
)

type Client struct {
	addr *net.TCPAddr
	name string

	lock        sync.Mutex
	c           *net.TCPConn
	isConnected bool
	isClosed    bool

	readWriteTimeout time.Duration
}

func NewClient(addr *net.TCPAddr, name string, timeout time.Duration) (c *Client) {
	c = &Client{
		addr:             addr,
		name:             name,
		readWriteTimeout: timeout,
	}

	return
}

func (c *Client) IsConnected() bool { return c.isConnected }
func (c *Client) IsClosed() bool    { return c.isClosed }

func (c *Client) Connect() (err error) {
	c.isClosed = false
	c.c, err = net.DialTCP("tcp", nil, c.addr)
	if err != nil {
		c.isConnected = false
		return
	}
	c.isConnected = true
	return
}

func (c *Client) Close() (err error) {
	c.isClosed = true
	c.isConnected = false
	err = c.c.Close()
	return
}

func (c *Client) GetId() string {
	return c.name
}

func (c *Client) DefaultAddressSpace(context.Context) (sni.AddressSpace, error) {
	return defaultAddressSpace, nil
}

func (c *Client) writeWithDeadline(bytes []byte, deadline time.Time) (err error) {
	err = c.c.SetWriteDeadline(deadline)
	if err != nil {
		return
	}
	_, err = c.c.Write(bytes)
	if err != nil {
		_ = c.Close()
		return
	}
	return
}

func (c *Client) readWithDeadline(b []byte, deadline time.Time) (n int, err error) {
	err = c.c.SetReadDeadline(deadline)
	if err != nil {
		return
	}
	n, err = c.c.Read(b)
	if err != nil {
		_ = c.Close()
		return
	}
	return
}

func (c *Client) SendCommandWaitReply(cmd string, deadline time.Time) (bin []byte, ascii []map[string]string, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	b := bytes.NewBuffer(make([]byte, 0, len(cmd)+1))
	b.WriteString(cmd)
	b.WriteByte('\n')
	err = c.writeWithDeadline(b.Bytes(), deadline)
	if err != nil {
		return
	}

	bin, ascii, err = c.parseCommandResponse(deadline)
	if ascii != nil {
		if errText, ok := ascii[0]["error"]; ok {
			err = fmt.Errorf("emunw: error=%s", errText)
			return
		}
	}
	return
}

func (c *Client) parseCommandResponse(deadline time.Time) (bin []byte, ascii []map[string]string, err error) {
	var n int
	// TODO: hope that packets Read() don't contain multiple responses
	b := make([]byte, 65536)
	n, err = c.readWithDeadline(b, deadline)
	if err != nil {
		_ = c.Close()
		return
	}

	d := b[:n]
	// parse binary reply:
	if d[0] == 0 {
		size := binary.BigEndian.Uint32(d[1 : 1+4])
		bin = b[5 : 5+size]
		return
	}
	// expect ascii reply otherwise:
	if d[0] != '\n' {
		err = fmt.Errorf("emunw: command reply expected starting with '\\0' or '\\n' but got '%c'", d[0])
		_ = c.Close()
		return
	}

	// parse ascii reply as array<map<string,string>>:
	s := bufio.NewScanner(bytes.NewReader(d[1:]))
	ascii = make([]map[string]string, 0, 4)
	item := make(map[string]string)
	for s.Scan() {
		l := s.Text()
		if l == "" {
			break
		}
		pair := strings.SplitN(l, ":", 2)
		var key string = pair[0]
		var value string
		if len(pair) >= 2 {
			value = pair[1]
		}

		// duplicate keys delimit multiple items:
		if _, hasKey := item[key]; hasKey {
			ascii = append(ascii, item)
			item = make(map[string]string)
		}

		item[key] = value
	}
	if len(item) > 0 {
		ascii = append(ascii, item)
	}

	return
}

type memRegion struct {
	mapping.MemoryType
	Offset uint32
	Size   int
	Data   []byte
}

func (c *Client) MultiReadMemory(ctx context.Context, reads ...snes.MemoryReadRequest) (mrsp []snes.MemoryReadResponse, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	mrsp = make([]snes.MemoryReadResponse, len(reads))

	// annoyingly, we must track the unique memType keys so we can iterate the map in a consistent order:
	memTypes := make([]mapping.MemoryType, 0, len(reads))
	readGroups := make(map[mapping.MemoryType][]memRegion)

	// divide up the reads into memory type groups:
	for j, read := range reads {
		a := &read.RequestAddress
		memType, pakAddress, offset := mapping.MemoryTypeFor(a)

		mrsp[j].RequestAddress = read.RequestAddress
		mrsp[j].DeviceAddress = snes.AddressTuple{
			Address:       pakAddress,
			AddressSpace:  sni.AddressSpace_FxPakPro,
			MemoryMapping: read.RequestAddress.MemoryMapping,
		}
		mrsp[j].DeviceAddress.Address = pakAddress
		mrsp[j].Data = make([]byte, read.Size)

		regions, ok := readGroups[memType]
		if !ok {
			memTypes = append(memTypes, memType)
		}
		readGroups[memType] = append(regions, memRegion{
			MemoryType: memType,
			Offset:     offset,
			Size:       read.Size,
			Data:       mrsp[j].Data,
		})
	}

	// write commands:
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, memType := range memTypes {
		regions := readGroups[memType]
		sb := bytes.Buffer{}
		_, _ = fmt.Fprintf(&sb, "CORE_READ %s", memType)
		for _, region := range regions {
			_, _ = fmt.Fprintf(&sb, ";$%x;$%x", region.Offset, region.Size)
		}
		sb.WriteByte('\n')
		err = c.writeWithDeadline(sb.Bytes(), deadline)
		if err != nil {
			return
		}
	}

	// read back responses:
	for _, memType := range memTypes {
		var bin []byte
		var ascii []map[string]string
		bin, ascii, err = c.parseCommandResponse(deadline)
		if err != nil {
			return
		}
		if ascii != nil {
			err = fmt.Errorf("emunw: expecting binary reply but got ascii:\n%+v", ascii)
		}

		regions := readGroups[memType]
		offset := 0
		for _, region := range regions {
			copy(region.Data, bin[offset:offset+region.Size])
			offset += region.Size
		}
	}

	err = nil
	return
}

func (c *Client) MultiWriteMemory(ctx context.Context, writes ...snes.MemoryWriteRequest) (mrsp []snes.MemoryWriteResponse, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	mrsp = make([]snes.MemoryWriteResponse, len(writes))

	// annoyingly, we must track the unique memType keys so we can iterate the map in a consistent order:
	memTypes := make([]mapping.MemoryType, 0, len(writes))
	writeGroups := make(map[mapping.MemoryType][]memRegion)

	// divide up the writes into memory type groups:
	for j, write := range writes {
		a := &write.RequestAddress
		memType, pakAddress, offset := mapping.MemoryTypeFor(a)

		mrsp[j].RequestAddress = write.RequestAddress
		mrsp[j].DeviceAddress = snes.AddressTuple{
			Address:       pakAddress,
			AddressSpace:  sni.AddressSpace_FxPakPro,
			MemoryMapping: write.RequestAddress.MemoryMapping,
		}
		mrsp[j].DeviceAddress.Address = pakAddress
		mrsp[j].Size = len(write.Data)

		regions, ok := writeGroups[memType]
		if !ok {
			memTypes = append(memTypes, memType)
		}
		writeGroups[memType] = append(regions, memRegion{
			MemoryType: memType,
			Offset:     offset,
			Size:       len(write.Data),
			Data:       write.Data,
		})
	}

	// write commands:
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, memType := range memTypes {
		regions := writeGroups[memType]

		// write command and build data buffer to send:
		sb := bytes.Buffer{}
		data := bytes.Buffer{}
		size := uint32(0)
		_, _ = fmt.Fprintf(&sb, "CORE_WRITE %s", memType)
		for _, region := range regions {
			_, _ = fmt.Fprintf(&sb, ";$%x;$%x", region.Offset, region.Size)
			data.Write(region.Data)
			size += uint32(region.Size)
		}
		sb.WriteByte('\n')
		err = c.writeWithDeadline(sb.Bytes(), deadline)
		if err != nil {
			return
		}

		// write data:
		sb.Reset()
		sb.WriteByte(0)
		_ = binary.Write(&sb, binary.BigEndian, size)
		sb.Write(data.Bytes())
		err = c.writeWithDeadline(sb.Bytes(), deadline)
		if err != nil {
			return
		}
	}

	// read replies:
	errReplies := strings.Builder{}
	for range memTypes {
		var ascii []map[string]string
		_, ascii, err = c.parseCommandResponse(deadline)
		if err != nil {
			return
		}
		if ascii != nil {
			if errText, ok := ascii[0]["error"]; ok {
				errReplies.WriteString(errText)
				errReplies.WriteByte('\n')
			}
		}
	}

	if errReplies.Len() > 0 {
		err = fmt.Errorf("emunw: error=%s", errReplies.String())
		return
	}

	err = nil
	return
}

func (c *Client) ResetSystem(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	_, _, err = c.SendCommandWaitReply("EMU_RESET", deadline)
	return
}

func (c *Client) PauseUnpause(ctx context.Context, pausedState bool) (newState bool, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.readWriteTimeout)
	}

	newState = pausedState
	if pausedState {
		_, _, err = c.SendCommandWaitReply("EMU_PAUSE", deadline)
	} else {
		_, _, err = c.SendCommandWaitReply("EMU_RESUME", deadline)
	}

	return
}

func (c *Client) PauseToggle(context.Context) (err error) {
	return fmt.Errorf("capability unavailable")
}
