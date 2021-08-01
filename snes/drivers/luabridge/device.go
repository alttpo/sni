package luabridge

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Device struct {
	lock sync.Mutex
	c    *net.TCPConn

	deviceKey string

	isClosed bool

	clientName string
	version    string
	host       string
	isBizHawk  bool
}

func NewDevice(conn *net.TCPConn, key string) *Device {
	d := &Device{
		c:          conn,
		deviceKey:  key,
		isClosed:   false,
		clientName: "Unknown",
		version:    "0",
	}
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

	log.Printf("luabridge: client '%s' version '%s' host '%s' bizhawk: %v\n", d.clientName, d.version, d.host, d.isBizHawk)

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

	// Version|SNI Connector|2|Bizhawk-bsnes
	// Version|SNI Connector|2|Bizhawk-snes9x
	// Version|SNI Connector|2|Snes9x
	d.clientName = rspn[1]
	d.version = rspn[2]
	if len(rspn) >= 4 {
		d.host = strings.ToLower(rspn[3])
		d.isBizHawk = strings.HasPrefix(d.host, "bizhawk")
	} else {
		d.host = ""
		d.isBizHawk = false
	}
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
	driver.DeleteDevice(d.deviceKey)

	return
}

func (d *Device) IsClosed() bool { return d.isClosed }
