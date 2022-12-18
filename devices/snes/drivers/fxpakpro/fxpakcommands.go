package fxpakpro

import (
	"fmt"
	"go.bug.st/serial"
	"sni/devices"
	"sync"
)

type fxpakCommands struct {
	lock sync.Mutex
	f    serial.Port
}

func (d *fxpakCommands) Close() error {
	return d.f.Close()
}

func (d *fxpakCommands) FatalError(cause error) devices.DeviceError {
	return devices.DeviceFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *fxpakCommands) NonFatalError(cause error) devices.DeviceError {
	return devices.DeviceNonFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}
