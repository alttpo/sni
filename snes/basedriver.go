package snes

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
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

func (b *BaseDeviceDriver) Put(deviceKey string, device Device) {
	b.devicesRw.Lock()
	b.devicesMap[deviceKey] = device
	b.devicesRw.Unlock()
}

func CheckCapabilities(expectedCapabilities []sni.DeviceCapability, actualCapabilities []sni.DeviceCapability) (bool, error) {
	for _, expected := range expectedCapabilities {
		found := false
		for _, actual := range actualCapabilities {
			if expected == actual {
				found = true
				break
			}
		}
		if !found {
			return false, fmt.Errorf("missing required capability %s", sni.DeviceCapability_name[int32(expected)])
		}
	}
	return true, nil
}

type BaseDeviceMemory struct {
	DeviceMemory
	Mapping sni.MemoryMapping
}

func (c *BaseDeviceMemory) MappingDetect(
	ctx context.Context,
	fallbackMapping *sni.MemoryMapping,
	inHeaderBytes []byte,
) (rsp sni.MemoryMapping, confidence bool, outHeaderBytes []byte, err error) {
	if inHeaderBytes == nil {
		var responses []MemoryReadResponse
		readRequest := MemoryReadRequest{
			RequestAddress:      0x00FFB0,
			RequestAddressSpace: sni.AddressSpace_SnesABus,
			Size:                0x50,
		}
		log.Printf(
			"detect: read {address:%s($%06x),size:$%x}\n",
			sni.AddressSpace_name[int32(readRequest.RequestAddressSpace)],
			readRequest.RequestAddress,
			readRequest.Size,
		)
		responses, err = c.DeviceMemory.MultiReadMemory(ctx, readRequest)
		if err != nil {
			return
		}

		outHeaderBytes = responses[0].Data

		log.Printf(
			"detect: read {address:%s($%06x),size:$%x} complete:\n%s",
			sni.AddressSpace_name[int32(responses[0].DeviceAddressSpace)],
			responses[0].DeviceAddress,
			len(outHeaderBytes),
			hex.Dump(outHeaderBytes),
		)
	} else {
		if len(inHeaderBytes) < 0x30 {
			err = fmt.Errorf("input ROM header must be at least $30 bytes")
			return
		}
		outHeaderBytes = inHeaderBytes
		log.Printf(
			"detect: provided header bytes {size:$%x}:\n%s",
			len(outHeaderBytes),
			hex.Dump(outHeaderBytes),
		)
	}

	header := Header{}
	err = header.ReadHeader(bytes.NewReader(outHeaderBytes))
	if err != nil {
		return
	}

	// detection does not have to be perfect (and never could be) since the client
	// always has the ability to override it or not use it at all and set their own
	// memory mapping.

	log.Printf(
		"detect: map mode %02x\n",
		header.MapMode&0b1110_1111,
	)

	confidence = true

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
		confidence = false
		if fallbackMapping != nil {
			c.Mapping = *fallbackMapping
			log.Printf(
				"detect: unable to detect mapping mode; falling back to provided default %s\n",
				sni.MemoryMapping_name[int32(c.Mapping)],
			)
		} else {
			// revert to a simple LoROM vs HiROM:
			c.Mapping = sni.MemoryMapping_LoROM - sni.MemoryMapping(header.MapMode&1)
			log.Printf(
				"detect: unable to detect mapping mode; guessing %s\n",
				sni.MemoryMapping_name[int32(c.Mapping)],
			)
		}
	}

	if confidence {
		log.Printf(
			"detect: detected mapping mode = %s\n",
			sni.MemoryMapping_name[int32(c.Mapping)],
		)
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
