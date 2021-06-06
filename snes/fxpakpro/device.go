package fxpakpro

import (
	"fmt"
	"sni/snes"
)

type DeviceDescriptor struct {
	snes.DeviceDescriptorBase
	Port string `json:"port"`
	Baud *int   `json:"baud"`
	VID  string `json:"vid"`
	PID  string `json:"pid"`
}

func (d *DeviceDescriptor) Base() *snes.DeviceDescriptorBase {
	return &d.DeviceDescriptorBase
}

func (d *DeviceDescriptor) GetId() string { return d.Port }

func (d *DeviceDescriptor) GetDisplayName() string {
	return fmt.Sprintf("%s (%s:%s)", d.Port, d.VID, d.PID)
}
