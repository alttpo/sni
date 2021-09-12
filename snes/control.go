package snes

import (
	"context"
)

type DeviceControl interface {
	ResetSystem(ctx context.Context) error
	ResetToMenu(ctx context.Context) error

	PauseUnpause(ctx context.Context, pausedState bool) (bool, error)
	PauseToggle(ctx context.Context) error
}
