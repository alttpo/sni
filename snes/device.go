package snes

import "io"

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	io.Closer
	DeviceControl
	DeviceMemory

	IsClosed() bool
}
