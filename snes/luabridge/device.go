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
	snes.BaseDeviceMemory

	lock sync.Mutex
	c    *net.TCPConn

	driver    *Driver
	deviceKey string

	isClosed bool

	clientName string
	version    string
}

func NewDevice(conn *net.TCPConn, key string, driver *Driver) *Device {
	d := &Device{
		c:          conn,
		driver:     driver,
		deviceKey:  key,
		isClosed:   false,
		clientName: "Unknown",
		version:    "0",
	}
	d.DeviceMemory = d
	return d
}

func (d *Device) Init() {
	go d.initConnection()
}

func (d *Device) initConnection() {
	var err error
	defer func() {
		if err != nil {
			log.Printf("luabridge: %v\n", err)
			err := d.Close()
			if err != nil {
				log.Printf("luabridge: close error: %v\n", err)
				return
			}
		}
	}()

	_ = d.c.SetNoDelay(true)

	err = d.CheckVersion()
	if err != nil {
		return
	}

	log.Printf("luabridge: client '%s' version '%s'\n", d.clientName, d.version)

	for {
		// every 5 seconds, check if connection is closed; only way to do this reliably is to read data:
		time.Sleep(time.Second * 5)

		err = d.CheckVersion()
		if err != nil {
			return
		}
	}
}

func (d *Device) CheckVersion() (err error) {
	b := make([]byte, 65535)

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

	d.clientName = rspn[1]
	d.version = rspn[2]
	return
}

func (d *Device) WriteDeadline(write []byte, deadline time.Time) (n int, err error) {
	defer d.lock.Unlock()
	d.lock.Lock()

	err = d.c.SetWriteDeadline(deadline)
	if err != nil {
		return
	}

	n, err = d.c.Write(write)
	if err != nil {
		return
	}

	return
}

func (d *Device) ReadDeadline(read []byte, deadline time.Time) (n int, err error) {
	defer d.lock.Unlock()
	d.lock.Lock()

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

	// remove device from driver:
	d.driver.devicesRw.Lock()
	delete(d.driver.devicesMap, d.deviceKey)
	d.driver.devicesRw.Unlock()

	return
}

func (d *Device) IsClosed() bool { return d.isClosed }

func (d *Device) Use(ctx context.Context, user snes.DeviceUser) error {
	return user(ctx, d)
}

func (d *Device) UseMemory(ctx context.Context, user snes.DeviceMemoryUser) error {
	return user(ctx, d)
}

func (d *Device) UseControl(ctx context.Context, user snes.DeviceControlUser) error {
	return user(ctx, d)
}
