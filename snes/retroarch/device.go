package retroarch

import (
	"context"
	"fmt"
	"net"
	"sni/snes"
	"sync"
)

type DeviceDescriptor struct {
	snes.DeviceDescriptorBase

	addr *net.UDPAddr

	IsGameLoaded bool `json:"isGameLoaded"`
}

func (d *DeviceDescriptor) Base() *snes.DeviceDescriptorBase {
	return &d.DeviceDescriptorBase
}

func (d *DeviceDescriptor) GetId() string {
	// dirty hack to work with JSON unmarshaled descriptors which won't have `addr` coming back:
	if d.addr == nil {
		return d.Id
	}
	return d.addr.String()
}

func (d *DeviceDescriptor) GetDisplayName() string {
	return fmt.Sprintf("RetroArch at %s", d.addr)
}

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
