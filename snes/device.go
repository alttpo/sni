package snes

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	UseControl
	UseMemory

	IsClosed() bool
}
