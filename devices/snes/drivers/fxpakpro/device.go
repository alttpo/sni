package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"sni/devices"
	"sync"
	"time"
)

type Device struct {
	lock sync.Mutex
	f    serial.Port

	isClosed bool
}

func (d *Device) FatalError(cause error) devices.DeviceError {
	return devices.DeviceFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *Device) NonFatalError(cause error) devices.DeviceError {
	return devices.DeviceNonFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *Device) Init() (err error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	// run an INFO request to make sure the fxpakpro is in a valid state, else this should
	// hopefully be a self-healing close/open loop:
	var version, device, rom string
	version, device, rom, err = d.info(ctx)
	if err != nil {
		return
	}
	if len(version) == 0 {
		err = d.FatalError(fmt.Errorf("bad INFO response; device version is empty"))
		return
	}
	if len(device) == 0 {
		err = d.FatalError(fmt.Errorf("bad INFO response; device name is empty"))
		return
	}
	if len(rom) == 0 {
		err = d.FatalError(fmt.Errorf("bad INFO response; rom name is empty"))
		return
	}

	return
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
