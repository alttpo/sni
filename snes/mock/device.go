package mock

import "sni/snes"

type DeviceDescriptor struct {
	snes.DeviceDescriptorBase
}

func (d *DeviceDescriptor) Base() *snes.DeviceDescriptorBase {
	return &d.DeviceDescriptorBase
}

func (d *DeviceDescriptor) GetId() string {
	return "mock"
}

func (d *DeviceDescriptor) GetDisplayName() string {
	return "Mock"
}
