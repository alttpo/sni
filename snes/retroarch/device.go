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
	// determine version:
	return d.c.DetermineVersion()
}

func (d *Device) IsClosed() bool {
	return d.c.IsClosed()
}

func (d *Device) UseMemory(context context.Context, user snes.MemoryUser) error {
	if user == nil {
		return nil
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	return user(context, d.c)
}
