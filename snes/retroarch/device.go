package retroarch

import (
	"context"
	"sni/protos/sni"
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

func (d *Device) UseMemory(ctx context.Context, requiredCapabilities []sni.DeviceCapability, user snes.DeviceMemoryUser) (err error) {
	if user == nil {
		return nil
	}

	if ok, err := driver.HasCapabilities(requiredCapabilities...); !ok {
		return err
	}

	//defer d.lock.Unlock()
	//d.lock.Lock()

	return user(ctx, d.c)
}

func (d *Device) UseControl(ctx context.Context, requiredCapabilities []sni.DeviceCapability, user snes.DeviceControlUser) error {
	if user == nil {
		return nil
	}

	if ok, err := driver.HasCapabilities(requiredCapabilities...); !ok {
		return err
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	return user(ctx, d.c)
}
