package luabridge

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sni/devices"
	"sni/protos/sni"
	"sni/util"
	"strings"
	"sync"
	"time"
)

type Device struct {
	stateLock sync.Mutex

	lock sync.Mutex
	c    *net.TCPConn

	deviceKey string

	isClosed bool

	lineReader *bufio.Reader

	clientName string
	version    string
	host       string
	isBizHawk  bool
	logPrefix  string
}

func (d *Device) FatalError(cause error) devices.DeviceError {
	return devices.DeviceFatal(fmt.Sprintf("%v", cause), cause)
}

func (d *Device) NonFatalError(cause error) devices.DeviceError {
	return devices.DeviceNonFatal(fmt.Sprintf("%v", cause), cause)
}

func NewDevice(conn *net.TCPConn, key string) *Device {
	d := &Device{
		c:          conn,
		deviceKey:  key,
		isClosed:   false,
		clientName: "Unknown",
		version:    "0",
		logPrefix:  fmt.Sprintf("luabridge:%s: ", conn.RemoteAddr()),
	}
	d.lineReader = bufio.NewReaderSize(d.c, 65536)
	return d
}

func (d *Device) log(f string, args ...interface{}) {
	log.Printf(d.logPrefix+f, args...)
}

func (d *Device) Init() {
	go d.initConnection()
}

func (d *Device) initConnection() {
	defer util.Recover()

	var err error
	remoteAddr := d.c.RemoteAddr()

	defer func() {
		if err != nil {
			d.log("%v\n", err)
		}

		err := d.Close()
		if err != nil {
			d.log("close error: %v\n", err)
		}

		d.log("connection closed")
		driver.DeleteDevice(remoteAddr.String())
	}()

	_ = d.c.SetNoDelay(true)

	// every 5 seconds, check if connection is closed; only way to do this reliably is to read data:
	first := true
	for {
		err = d.CheckVersion()
		if err != nil {
			d.log("error while checking version")
			return
		}

		if first {
			d.log("client '%s' version '%s' host '%s' bizhawk: %v\n", d.clientName, d.version, d.host, d.isBizHawk)
			first = false
		}

		time.Sleep(time.Second * 5)
	}
}

func (d *Device) CheckVersion() (err error) {
	defer d.stateLock.Unlock()
	d.stateLock.Lock()

	var b []byte
	b, err = d.WriteThenReadUntilNewline([]byte("Version\n"), time.Now().Add(time.Second*15))
	if err != nil {
		return
	}

	rsp := string(b[:])
	rsp = strings.TrimRight(rsp, "\r\n ")

	rspn := strings.Split(rsp, "|")
	if len(rspn) < 3 {
		err = fmt.Errorf("expected Version response")
		err = d.FatalError(err)
		return
	}
	if rspn[0] != "Version" {
		err = fmt.Errorf("expected Version response")
		err = d.FatalError(err)
		return
	}

	// Version|SNI Connector|2|Bizhawk-bsnes
	// Version|SNI Connector|2|Bizhawk-snes9x
	// Version|SNI Connector|2|Snes9x
	// Version|SNI Connector|3|Snes9x
	// Version|SNI Connector|4|Snes9x
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

func (d *Device) FetchFields(ctx context.Context, fields ...sni.Field) (values []string, err error) {
	for _, field := range fields {
		switch field {
		case sni.Field_DeviceName:
			values = append(values, d.clientName+"|"+d.host)
			break
		case sni.Field_DeviceVersion:
			values = append(values, d.version)
			break
		default:
			// unknown value; append empty string to maintain index association:
			values = append(values, "")
			break
		}
	}

	return
}

func (d *Device) writeUnderLock(write []byte, deadline time.Time) (n int, err error) {
	err = d.c.SetWriteDeadline(deadline)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	n, err = d.c.Write(write)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	return
}

func (d *Device) readUntilNewlineUnderLock(deadline time.Time) (line []byte, err error) {
	err = d.c.SetReadDeadline(deadline)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	line, err = d.lineReader.ReadBytes('\n')
	if err != nil {
		err = d.FatalError(err)
		return
	}

	return
}

func (d *Device) WriteDeadline(write []byte, deadline time.Time) (n int, err error) {
	defer d.lock.Unlock()
	d.lock.Lock()

	return d.writeUnderLock(write, deadline)
}

func (d *Device) WriteThenReadUntilNewline(write []byte, deadline time.Time) (line []byte, err error) {
	defer d.lock.Unlock()
	d.lock.Lock()

	if _, err = d.writeUnderLock(write, deadline); err != nil {
		return
	}

	return d.readUntilNewlineUnderLock(deadline)
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
