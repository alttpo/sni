package fxpakpro

import (
	"fmt"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
	"log"
	"sni/snes"
	"sni/util"
	"sni/util/env"
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

func (d *Driver) Empty() snes.DeviceDescriptor {
	return &DeviceDescriptor{
		Port: "",
		Baud: &(baudRates[0]),
	}
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

		//log.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
		//log.Printf("   USB serial %s\n", port.SerialNumber)

		if port.SerialNumber == "DEMO00000000" {
			devices = append(devices, &DeviceDescriptor{
				snes.DeviceDescriptorBase{
					Id:          port.Name,
					DisplayName: port.Name,
				},
				port.Name,
				nil,
				port.VID,
				port.PID,
			})
		}
	}

	err = nil
	return
}

func (d *Driver) Open(ddg snes.DeviceDescriptor) (snes.Queue, error) {
	var err error

	dd := ddg.(*DeviceDescriptor)
	portName := dd.Port
	if portName == "" {
		ddgs, err := d.Detect()
		if err != nil {
			return nil, err
		}

		// pick first device found, if any:
		if len(ddgs) > 0 {
			portName = ddgs[0].(*DeviceDescriptor).Port
		}
	}
	if portName == "" {
		return nil, ErrNoFXPakProFound
	}

	baudRequest := baudRates[0]
	if dd.Baud != nil {
		b := *dd.Baud
		if b > 0 {
			baudRequest = b
		}
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

	// set baud rate on descriptor:
	pBaud := new(int)
	*pBaud = baud
	dd.Baud = pBaud

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

func init() {
	if util.IsTruthy(env.GetOrDefault("SNI_FXPAKPRO_DISABLE", "0")) {
		log.Printf("disabling fxpakpro snes driver\n")
		return
	}
	snes.Register(driverName, &Driver{})
}
