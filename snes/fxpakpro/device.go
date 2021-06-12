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

// TODO: algorithm to break up reads into VGET 255-byte chunks:
//addr := request.GetAddress()
//size := int32(request.GetSize())
//reads := make([]snes.Read, 0, 8)
//data := make([]byte, 0, size)
//for size > 0 {
//chunkSize := int32(255)
//if size < chunkSize {
//chunkSize = size
//}
//
//reads = append(reads, snes.Read{
//Address: addr,
//Size:    uint8(chunkSize),
//Extra:   nil,
//Completion: func(response snes.Response) {
//data = append(data, response.Data...)
//},
//})
//
//size -= 255
//addr += 255
//}
