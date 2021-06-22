package snes

import "context"

type DeviceControlUser func(ctx context.Context, control DeviceControl) error

type DeviceControl interface {
	ResetSystem(ctx context.Context) error

	PauseUnpause(ctx context.Context, pausedState bool) (bool, error)
	PauseToggle(ctx context.Context) error
}
