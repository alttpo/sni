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

type Driver struct{}

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
				Uri:          url.URL{Scheme: driverName, Path: port.Name},
				DisplayName:  fmt.Sprintf("%s (%s:%s)", port.Name, port.VID, port.PID),
				Kind:         d.Kind(),
				Capabilities: sni.DeviceCapability_READ | sni.DeviceCapability_WRITE | sni.DeviceCapability_EXEC_ASM | sni.DeviceCapability_RESET,
			})
		}
	}

	err = nil
	return
}

func (d *Driver) OpenQueue(dd snes.DeviceDescriptor) (snes.Queue, error) {
	var err error

	portName := dd.Uri.Path

	baudRequest := baudRates[0]
	if baudStr := dd.Uri.Query().Get("baud"); baudStr != "" {
		baudRequest, _ = strconv.Atoi(baudStr)
	}

	// Try all the common baud rates in descending order:
	f := serial.Port(nil)
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
		f.Close()
		return nil, fmt.Errorf("%s: failed to set DTR: %w", driverName, err)
	}

	c := &Queue{
		f:      f,
		closed: make(chan struct{}),
	}
	c.BaseInit(driverName, c)

	return c, err
}

func (d *Driver) OpenDevice(uri *url.URL) (snes.Device, error) {
	panic("implement me")
}

func (d *Driver) UseDevice(ctx context.Context, uri *url.URL, user snes.DeviceUser) error {
	panic("implement me")
}

func init() {
	if util.IsTruthy(env.GetOrDefault("SNI_FXPAKPRO_DISABLE", "0")) {
		log.Printf("disabling fxpakpro snes driver\n")
		return
	}
	snes.Register(driverName, &Driver{})
}
