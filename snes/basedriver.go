package snes

import (
	"fmt"
	"net/url"
	"sni/protos/sni"
	"sync"
)

type DeviceDriverContainer interface {
	GetDevice(deviceKey string) (Device, bool)
	PutDevice(deviceKey string, device Device)
	DeleteDevice(deviceKey string)
	OpenDevice(deviceKey string, uri *url.URL, opener DeviceOpener) (device Device, err error)
}

type BaseDeviceDriver struct {
	// track opened devices by URI
	devicesRw  sync.RWMutex
	devicesMap map[string]Device
}

func (b *BaseDeviceDriver) GetDevice(deviceKey string) (Device, bool) {
	b.devicesRw.RLock()
	device, ok := b.devicesMap[deviceKey]
	b.devicesRw.RUnlock()

	return device, ok
}

func (b *BaseDeviceDriver) PutDevice(deviceKey string, device Device) {
	b.devicesRw.Lock()
	b.devicesMap[deviceKey] = device
	b.devicesRw.Unlock()
}

func (b *BaseDeviceDriver) DeleteDevice(deviceKey string) {
	b.devicesRw.Lock()
	b.deleteUnderLock(deviceKey)
	b.devicesRw.Unlock()
}

func (b *BaseDeviceDriver) deleteUnderLock(deviceKey string) {
	if b.devicesMap == nil {
		b.devicesMap = make(map[string]Device)
	}
	delete(b.devicesMap, deviceKey)
}

func (b *BaseDeviceDriver) OpenDevice(deviceKey string, uri *url.URL, opener DeviceOpener) (device Device, err error) {
	b.devicesRw.Lock()
	device, err = opener(uri)
	if err != nil {
		b.deleteUnderLock(deviceKey)
		b.devicesRw.Unlock()
		return
	}

	if b.devicesMap == nil {
		b.devicesMap = make(map[string]Device)
	}
	b.devicesMap[deviceKey] = device
	b.devicesRw.Unlock()
	return
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
