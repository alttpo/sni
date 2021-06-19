package luabridge

import (
	"context"
	"fmt"
	"log"
	"net"
	"sni/snes"
	"strings"
	"sync"
	"time"
)

type Device struct {
	lock sync.Mutex
	c    *net.TCPConn

	deviceKey string

	isClosed bool
	onClose  func(device *Device)
}

func (d *Device) handleConnection() {
	var err error
	defer func() {
		if err != nil {
			log.Printf("luabridge: %v\n", err)
		}

		err := d.Close()
		if err != nil {
			log.Printf("luabridge: close error: %v\n", err)
			return
		}
	}()

	b := make([]byte, 65535)

	_ = d.c.SetNoDelay(true)

	var n int
	n, err = d.WriteThenRead([]byte("Version\n"), b, time.Now().Add(time.Second*15))
	if err != nil {
		return
	}

	rsp := string(b[:n])
	rsp = strings.TrimRight(rsp, "\r\n ")

	rspn := strings.Split(rsp, "|")
	if len(rspn) < 3 {
		err = fmt.Errorf("expected Version response")
		return
	}
	if rspn[0] != "Version" {
		err = fmt.Errorf("expected Version response")
		return
	}
	client := rspn[1]
	version := rspn[2]

	log.Printf("luabridge: client '%s' version '%s'\n", client, version)

	// TODO: read/write loop
}

func (d *Device) WriteThenRead(write []byte, read []byte, deadline time.Time) (n int, err error) {
	defer d.lock.Unlock()
	d.lock.Lock()

	err = d.c.SetWriteDeadline(deadline)
	if err != nil {
		return
	}

	_, err = d.c.Write(write)
	if err != nil {
		return
	}

	err = d.c.SetReadDeadline(deadline)
	if err != nil {
		return
	}

	n, err = d.c.Read(read)
	if err != nil {
		return
	}

	return
}

func (d *Device) Close() (err error) {
	if d.isClosed {
		return nil
	}

	d.isClosed = true
	err = d.c.Close()
	d.onClose(d)
	return
}

func (d *Device) IsClosed() bool { return d.isClosed }

func (d *Device) Use(ctx context.Context, user snes.DeviceUser) error {
	panic("implement me")
}

func (d *Device) UseMemory(ctx context.Context, user snes.DeviceMemoryUser) error {
	panic("implement me")
}

func (d *Device) UseControl(ctx context.Context, user snes.DeviceControlUser) error {
	panic("implement me")
}
