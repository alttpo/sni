package snes

import (
	"context"
	"sni/protos/sni"
)

type DeviceControlUser func(ctx context.Context, control DeviceControl) error

type DeviceControl interface {
	ResetSystem(ctx context.Context) error

	PauseUnpause(ctx context.Context, pausedState bool) (bool, error)
	PauseToggle(ctx context.Context) error
}

type UseControl interface {
	// UseControl provides exclusive access to only the control subsystem of the device to the user func
	UseControl(ctx context.Context, requiredCapabilities []sni.DeviceCapability, user DeviceControlUser) error
}
