package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"sni/devices"
	"sync"
	"sync/atomic"
	"time"
)

// devicePort wraps the serial port so that any close marks the device closed --
// including one done by the write path, which closes the port when it abandons
// a write so an orphaned goroutine cannot interleave with later commands. That
// close does not go through Device.Close(), so without this the isClosed flag
// would stay false and autoCloseableDevice would keep a device whose port is
// dead in its container (it consults IsClosed() to decide whether to drop it).
type devicePort struct {
	serial.Port
	closed atomic.Bool
}

func (p *devicePort) Close() error {
	p.closed.Store(true)
	return p.Port.Close()
}

type Device struct {
	lock sync.Mutex
	f    *devicePort
}

func (d *Device) FatalError(cause error) devices.DeviceError {
	return devices.DeviceFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *Device) NonFatalError(cause error) devices.DeviceError {
	return devices.DeviceNonFatal(fmt.Sprintf("fxpakpro: %v", cause), cause)
}

func (d *Device) Init() (err error) {
	// This budget used to be a hardcoded 2 seconds, back when a context
	// deadline only bounded reads. It now bounds writes as well, and abandoning
	// a write closes the port -- so too tight a value here would close healthy
	// devices. A write only blocks when the device is not draining its USB
	// endpoint, which happens while the firmware is busy inside its interrupt
	// handler: FatFs cluster allocation on a large, full, fragmented card has
	// been measured stalling for hundreds of milliseconds and can reach seconds.
	// Use the same configured budget as every other read, so there is one knob
	// (fxpakpro_read_timeout) rather than a hidden second one.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(noDataTimeout))
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
	return d.f.closed.Load()
}

func (d *Device) Close() (err error) {
	return d.f.Close()
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
