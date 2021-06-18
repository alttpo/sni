package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
	"log"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
	"sni/util"
	"sni/util/env"
	"strconv"
)

const (
	driverName = "fxpakpro"
)

var (
	ErrNoFXPakProFound = fmt.Errorf("%s: no device found among serial ports", driverName)

	baudRates = []int{
		921600, // first rate that works on Windows
		460800,
		256000,
		230400, // first rate that works on MacOS
		153600,
		128000,
		115200,
		76800,
		57600,
		38400,
		28800,
		19200,
		14400,
		9600,
	}
)

type Driver struct {
	base snes.BaseDeviceDriver
}

func (d *Driver) DisplayOrder() int {
	return 0
}

func (d *Driver) DisplayName() string {
	return "FX Pak Pro"
}

func (d *Driver) DisplayDescription() string {
	return "Connect to an FX Pak Pro or SD2SNES via USB"
}

func (d *Driver) Kind() string { return "fxpakpro" }

var driverCapabilities = []sni.DeviceCapability{
	sni.DeviceCapability_ReadMemory,
	sni.DeviceCapability_WriteMemory,
	sni.DeviceCapability_ResetSystem,
	sni.DeviceCapability_ExecuteASM,
}

func (d *Driver) HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error) {
	return snes.CheckCapabilities(capabilities, driverCapabilities)
}

func (d *Driver) Detect() (devices []snes.DeviceDescriptor, err error) {
	var ports []*enumerator.PortDetails

	// It would be surprising to see more than one FX Pak Pro connected to a PC.
	devices = make([]snes.DeviceDescriptor, 0, 1)

	ports, err = enumerator.GetDetailedPortsList()
	if err != nil {
		return
	}

	for _, port := range ports {
		if !port.IsUSB {
			continue
		}

		if port.SerialNumber == "DEMO00000000" {
			devices = append(devices, snes.DeviceDescriptor{
				Uri:                 url.URL{Scheme: driverName, Path: port.Name},
				DisplayName:         fmt.Sprintf("%s (%s:%s)", port.Name, port.VID, port.PID),
				Kind:                d.Kind(),
				Capabilities:        driverCapabilities[:],
				DefaultAddressSpace: sni.AddressSpace_FxPakPro,
			})
		}
	}

	err = nil
	return
}

func (d *Driver) OpenQueue(dd snes.DeviceDescriptor) (q snes.Queue, err error) {
	portName := dd.Uri.Path

	baudRequest := baudRates[0]
	if baudStr := dd.Uri.Query().Get("baud"); baudStr != "" {
		baudRequest, _ = strconv.Atoi(baudStr)
	}

	var f serial.Port
	f, err = d.openPort(portName, baudRequest)
	if err != nil {
		return
	}

	c := &Queue{
		f:      f,
		closed: make(chan struct{}),
	}
	c.BaseInit(driverName, c)

	q = c

	return
}

func (d *Driver) openPort(portName string, baudRequest int) (f serial.Port, err error) {
	f = serial.Port(nil)

	// Try all the common baud rates in descending order:
	var baud int
	for _, baud = range baudRates {
		if baud > baudRequest {
			continue
		}

		//log.Printf("%s: open(%d)\n", portName, baud)
		f, err = serial.Open(portName, &serial.Mode{
			BaudRate: baud,
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		})
		if err == nil {
			break
		}
		//log.Printf("%s: %v\n", portName, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open serial port at any baud rate: %w", driverName, err)
	}

	// set DTR:
	//log.Printf("serial: Set DTR on\n")
	if err = f.SetDTR(true); err != nil {
		//log.Printf("serial: %v\n", err)
		_ = f.Close()
		return nil, fmt.Errorf("%s: failed to set DTR: %w", driverName, err)
	}

	return f, err
}

func (d *Driver) DeviceKey(uri *url.URL) string {
	return uri.Path
}

func (d *Driver) openAsDevice(uri *url.URL) (device snes.Device, err error) {
	portName := uri.Path

	baudRequest := baudRates[0]
	if baudStr := uri.Query().Get("baud"); baudStr != "" {
		baudRequest, _ = strconv.Atoi(baudStr)
	}

	var f serial.Port
	f, err = d.openPort(portName, baudRequest)
	dev := &Device{f: f}
	err = dev.Init()

	device = dev
	return
}

func (d *Driver) UseDevice(ctx context.Context, uri *url.URL, user snes.DeviceUser) error {
	return d.base.UseDevice(
		ctx,
		d.DeviceKey(uri),
		func() (snes.Device, error) {
			return d.openAsDevice(uri)
		},
		user,
	)
}

func init() {
	if util.IsTruthy(env.GetOrDefault("SNI_FXPAKPRO_DISABLE", "0")) {
		log.Printf("disabling fxpakpro snes driver\n")
		return
	}
	snes.Register(driverName, &Driver{})
}
