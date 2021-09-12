package luabridge

import (
	"context"
	"time"
)

func (d *Device) ResetSystem(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	_, err = d.WriteDeadline([]byte("Reset\n"), deadline)
	return
}

func (d *Device) ResetToMenu(ctx context.Context) error {
	panic("implement me")
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (paused bool, err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	paused = pausedState
	if paused {
		_, err = d.WriteDeadline([]byte("Pause\n"), deadline)
	} else {
		_, err = d.WriteDeadline([]byte("Unpause\n"), deadline)
	}
	return
}

func (d *Device) PauseToggle(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	_, err = d.WriteDeadline([]byte("PauseToggle\n"), deadline)
	return
}
