package fxpakpro

import (
	"go.bug.st/serial"
	"sync"
)

type Device struct {
	lock sync.Mutex
	f    serial.Port

	isClosed bool
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
