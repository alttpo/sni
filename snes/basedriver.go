package snes

import (
	"context"
	"sync"
)

type BaseDeviceDriver struct {
	driver DeviceDriver

	// track opened devices by URI
	devicesRw  sync.RWMutex
	devicesMap map[string]Device
}

func (b *BaseDeviceDriver) UseDevice(ctx context.Context, deviceKey string, openDevice func() (Device, error), use DeviceUser) (err error) {
	var device Device
	var ok bool

	b.devicesRw.RLock()
	device, ok = b.devicesMap[deviceKey]
	b.devicesRw.RUnlock()

	if !ok {
		device, err = openDevice()
		if err != nil {
			return
		}

		b.devicesRw.Lock()
		if b.devicesMap == nil {
			b.devicesMap = make(map[string]Device)
		}
		b.devicesMap[deviceKey] = device
		b.devicesRw.Unlock()
	}

	err = device.ExclusiveUse(ctx, use)

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
