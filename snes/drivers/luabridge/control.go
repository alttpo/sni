package luabridge

import "context"

func (d *Device) ResetSystem(ctx context.Context) error {
	panic("implement me")
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	panic("implement me")
}

func (d *Device) PauseToggle(ctx context.Context) error {
	panic("implement me")
}
