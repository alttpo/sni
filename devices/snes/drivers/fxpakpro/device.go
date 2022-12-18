package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"sync"
)

type Device struct {
	lock sync.Mutex
	c    commands

	isClosed bool
}

func testableDevice(f serial.Port) *Device {
	return &Device{c: &fxpakCommands{f: f}}
}

func (d *Device) Init() error {
	return nil
}

func (d *Device) IsClosed() bool {
	return d.isClosed
}

func (d *Device) Close() (err error) {
	err = d.c.Close()
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
