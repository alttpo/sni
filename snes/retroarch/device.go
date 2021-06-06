package retroarch

import (
	"fmt"
	"net"
	"sni/snes"
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
