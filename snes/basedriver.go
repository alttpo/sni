package snes

import (
	"context"
	"fmt"
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
