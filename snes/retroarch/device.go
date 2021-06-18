package retroarch

import (
	"context"
	"sni/snes"
	"sync"
)

type Device struct {
	lock sync.Mutex
	c    *RAClient
}

func (d *Device) Init() error {
	// determine retroarch version:
	return d.c.DetermineVersion()
}

func (d *Device) IsClosed() bool {
	return d.c.IsClosed()
}

func (d *Device) Use(ctx context.Context, user snes.DeviceUser) error {
	if user == nil {
		return nil
	}

	return user(ctx, d)
}

func (d *Device) UseMemory(ctx context.Context, user snes.DeviceMemoryUser) error {
	if user == nil {
		return nil
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	return user(ctx, d.c)
}
