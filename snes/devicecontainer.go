package snes

import (
	"fmt"
	"net/url"
	"sni/protos/sni"
	"sync"
)

type DeviceContainer interface {
	OpenDevice(deviceKey string, uri *url.URL) (device Device, err error)
	GetDevice(deviceKey string) (Device, bool)
	GetOrOpenDevice(deviceKey string, uri *url.URL) (device Device, err error)
	PutDevice(deviceKey string, device Device)
	DeleteDevice(deviceKey string)
	AllDeviceKeys() []string
}

type DeviceOpener func(uri *url.URL) (Device, error)

type deviceContainer struct {
	opener DeviceOpener

	// track opened devices by URI
	devicesRw  sync.RWMutex
	devicesMap map[string]Device
}

func NewDeviceDriverContainer(opener DeviceOpener) DeviceContainer {
	return &deviceContainer{
		opener:     opener,
		devicesRw:  sync.RWMutex{},
		devicesMap: make(map[string]Device),
	}
}

func (b *deviceContainer) GetOrOpenDevice(deviceKey string, uri *url.URL) (device Device, err error) {
	var ok bool
	device, ok = b.GetDevice(deviceKey)
	if !ok {
		device, err = b.OpenDevice(deviceKey, uri)
		if err != nil {
			return
		}
	}
	return
}

func (b *deviceContainer) GetDevice(deviceKey string) (Device, bool) {
	b.devicesRw.RLock()
	device, ok := b.devicesMap[deviceKey]
	b.devicesRw.RUnlock()

	return device, ok
}

func (b *deviceContainer) PutDevice(deviceKey string, device Device) {
	b.devicesRw.Lock()
	b.devicesMap[deviceKey] = device
	b.devicesRw.Unlock()
}

func (b *deviceContainer) DeleteDevice(deviceKey string) {
	b.devicesRw.Lock()
	b.deleteUnderLock(deviceKey)
	b.devicesRw.Unlock()
}

func (b *deviceContainer) deleteUnderLock(deviceKey string) {
	if b.devicesMap == nil {
		b.devicesMap = make(map[string]Device)
	}
	delete(b.devicesMap, deviceKey)
}

func (b *deviceContainer) OpenDevice(deviceKey string, uri *url.URL) (device Device, err error) {
	b.devicesRw.Lock()
	device, err = b.opener(uri)
	if err != nil {
		b.deleteUnderLock(deviceKey)
		b.devicesRw.Unlock()
		return
	}

	b.devicesMap[deviceKey] = device
	b.devicesRw.Unlock()
	return
}

func (b *deviceContainer) AllDeviceKeys() []string {
	defer b.devicesRw.RUnlock()
	b.devicesRw.RLock()
	deviceKeys := make([]string, 0, len(b.devicesMap))
	for deviceKey := range b.devicesMap {
		deviceKeys = append(deviceKeys, deviceKey)
	}
	return deviceKeys
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
