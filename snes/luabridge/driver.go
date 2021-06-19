package luabridge

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
	"sync"
)

const driverName = "luabridge"

type Driver struct {
	// track opened devices by URI
	devicesRw  sync.RWMutex
	devicesMap map[string]*Device
}

func (d *Driver) Kind() string {
	return "luabridge"
}

var driverCapabilities = []sni.DeviceCapability{
	sni.DeviceCapability_ReadMemory,
	sni.DeviceCapability_WriteMemory,
}

func (d *Driver) HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error) {
	return snes.CheckCapabilities(capabilities, driverCapabilities)
}

func (d *Driver) Detect() (devices []snes.DeviceDescriptor, err error) {
	d.devicesRw.RLock()
	devices = make([]snes.DeviceDescriptor, 0, len(d.devicesMap))
	for _, device := range d.devicesMap {
		devices = append(devices, snes.DeviceDescriptor{
			Uri:                 url.URL{Scheme: driverName, Host: device.c.RemoteAddr().String()},
			DisplayName:         fmt.Sprintf("%s v%s", device.clientName, device.version),
			Kind:                d.Kind(),
			Capabilities:        driverCapabilities[:],
			DefaultAddressSpace: sni.AddressSpace_SnesABus,
		})
	}
	d.devicesRw.RUnlock()
	return
}

func (d *Driver) DeviceKey(uri *url.URL) string {
	return uri.Host
}

func (d *Driver) UseDevice(ctx context.Context, uri *url.URL, user snes.DeviceUser) (err error) {
	deviceKey := d.DeviceKey(uri)

	d.devicesRw.RLock()
	device, ok := d.devicesMap[deviceKey]
	d.devicesRw.RUnlock()

	if !ok {
		return fmt.Errorf("no device found")
	}

	err = device.Use(ctx, user)

	return
}

func (d *Driver) StartServer() (err error) {
	d.devicesMap = make(map[string]*Device)

	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp4", "127.0.0.1:65398")
	if err != nil {
		return
	}

	var listener *net.TCPListener
	listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		return
	}

	go d.runServer(listener)

	return
}

func (d *Driver) runServer(listener *net.TCPListener) {
	var err error
	defer listener.Close()

	// TODO: stopping criteria
	for {
		// accept new TCP connections:
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			break
		}

		// create the Device to handle this connection:
		deviceKey := conn.RemoteAddr().String()
		device := NewDevice(conn, deviceKey, d)

		// store the Device for reference:
		d.devicesRw.Lock()
		d.devicesMap[deviceKey] = device
		d.devicesRw.Unlock()

		// initialize the Device:
		device.Init()
	}
}

func init() {
	// attempt to start the luabridge server:
	driver := &Driver{}
	err := driver.StartServer()
	if err != nil {
		log.Printf("luabridge: could not start server: %v\n", err)
		return
	}

	snes.Register(driverName, driver)
}
