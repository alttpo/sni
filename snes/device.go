package snes

import (
	"context"
)

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	IsClosed() bool

	// Use provides non-exclusive access to the device's subsystems to the user func
	Use(ctx context.Context, user DeviceUser) error

	// UseMemory provides exclusive access to only the memory subsystem of the device to the user func
	UseMemory(ctx context.Context, user DeviceMemoryUser) error

	// UseControl provides exclusive access to only the control subsystem of the device to the user func
	UseControl(ctx context.Context, user DeviceControlUser) error
}
