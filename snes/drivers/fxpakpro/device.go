package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"sni/snes"
	"sync"
)

type Device struct {
	lock sync.Mutex
	f    serial.Port

	isClosed bool
}

func (d *Device) FatalError(cause error) snes.DeviceError {
	return snes.DeviceFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *Device) NonFatalError(cause error) snes.DeviceError {
	return snes.DeviceNonFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *Device) Init() error {
	return nil
}

func (d *Device) IsClosed() bool {
	return d.isClosed
}

func (d *Device) Close() (err error) {
	err = d.f.Close()
	d.isClosed = true
	return
}

type lockedKeyType int

var lockedKey lockedKeyType

func shouldLock(ctx context.Context) bool {
	return ctx.Value(lockedKey) == nil
}

type fxpakproError uint8

func (f fxpakproError) Error() string {
	return fmt.Sprintf("fxpakpro responded with error code %d", f)
}
