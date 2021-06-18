package snes

import (
	"bytes"
	"context"
	"sni/protos/sni"
	"sync"
)

type BaseDeviceDriver struct {
	// track opened devices by URI
	devicesRw  sync.RWMutex
	devicesMap map[string]Device
}

func (b *BaseDeviceDriver) UseDevice(
	ctx context.Context,
	deviceKey string,
	openDevice func() (Device, error),
	use DeviceUser,
) (err error) {
	var device Device
	var ok bool

	b.devicesRw.RLock()
	device, ok = b.devicesMap[deviceKey]
	b.devicesRw.RUnlock()

	if !ok {
		b.devicesRw.Lock()
		device, err = openDevice()
		if err != nil {
			b.devicesRw.Unlock()
			return
		}

		if b.devicesMap == nil {
			b.devicesMap = make(map[string]Device)
		}
		b.devicesMap[deviceKey] = device
		b.devicesRw.Unlock()
	}

	err = device.Use(ctx, use)

	if device.IsClosed() {
		b.devicesRw.Lock()
		if b.devicesMap == nil {
			b.devicesMap = make(map[string]Device)
		}
		delete(b.devicesMap, deviceKey)
		b.devicesRw.Unlock()
	}

	return
}

func (b *BaseDeviceDriver) Get(deviceKey string) (Device, bool) {
	b.devicesRw.RLock()
	device, ok := b.devicesMap[deviceKey]
	b.devicesRw.RUnlock()

	return device, ok
}

type BaseDeviceMemory struct {
	DeviceMemory
	Mapping sni.MemoryMapping
}

func (c *BaseDeviceMemory) MappingDetect(ctx context.Context, fallbackMapping *sni.MemoryMapping) (rsp sni.MemoryMapping, err error) {
	var responses []MemoryReadResponse
	responses, err = c.DeviceMemory.MultiReadMemory(ctx, MemoryReadRequest{
		RequestAddress:      0x00FFB0,
		RequestAddressSpace: sni.AddressSpace_SnesABus,
		Size:                0x30,
	})
	if err != nil {
		return
	}

	header := Header{}
	err = header.ReadHeader(bytes.NewReader(responses[0].Data))
	if err != nil {
		return
	}

	// detection does not have to be perfect (and never could be) since the client
	// always has the ability to override it or not use it at all and set their own
	// memory mapping.

	// mask off SlowROM vs FastROM bit:
	switch header.MapMode & 0b1110_1111 {
	case 0x20: // LoROM
		c.Mapping = sni.MemoryMapping_LoROM
	case 0x21: // HiROM
		c.Mapping = sni.MemoryMapping_HiROM
	case 0x22: // ExLoROM
		c.Mapping = sni.MemoryMapping_LoROM
	case 0x23: // SA-1
		c.Mapping = sni.MemoryMapping_HiROM
	case 0x25: // ExHiROM
		c.Mapping = sni.MemoryMapping_ExHiROM
	default:
		if fallbackMapping != nil {
			c.Mapping = *fallbackMapping
		} else {
			// revert to a simple LoROM vs HiROM:
			c.Mapping = sni.MemoryMapping_LoROM - sni.MemoryMapping(header.MapMode&1)
		}
	}

	rsp = c.Mapping
	return
}

func (c *BaseDeviceMemory) MappingSet(mapping sni.MemoryMapping) sni.MemoryMapping {
	c.Mapping = mapping
	return c.Mapping
}

func (c *BaseDeviceMemory) MappingGet() sni.MemoryMapping {
	return c.Mapping
}
