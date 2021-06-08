package main

import (
	"context"
	"fmt"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
	"sync"
)

type devicesService struct {
	sni.UnimplementedDevicesServiceServer

	// track opened devices by URI
	devicesRw sync.RWMutex
	devices   map[string]snes.Queue
}

func (s *devicesService) ListDevices(ctx context.Context, request *sni.DevicesRequest) (*sni.DevicesResponse, error) {
	var kindPredicate func(kind string) bool
	if request.GetKinds() == nil {
		kindPredicate = func(kind string) bool { return true }
	} else {
		kindPredicate = func(kind string) bool {
			for _, k := range request.GetKinds() {
				if kind == k {
					return true
				}
			}
			return false
		}
	}

	devices := make([]*sni.DevicesResponse_Device, 0, 10)
	for _, driver := range snes.Drivers() {
		if !kindPredicate(driver.Name) {
			continue
		}

		descriptors, err := driver.Driver.Detect()
		if err != nil {
			return nil, err
		}
		for _, descriptor := range descriptors {
			devices = append(devices, &sni.DevicesResponse_Device{
				Uri:         fmt.Sprintf("%s://%s", driver.Name, descriptor.GetId()),
				DisplayName: descriptor.GetDisplayName(),
				Kind:        driver.Name,
				// TODO: get device version from descriptor:
				Version: "TODO", //descriptor.GetVersion(),
				// TODO: get capabilities from descriptor:
				Capabilities: int32(sni.DeviceCapability_READ | sni.DeviceCapability_WRITE), //descriptor.GetCapabilities(),
			})
		}
	}

	return &sni.DevicesResponse{Devices: devices}, nil
}

func (s *devicesService) ReadMemory(ctx context.Context, request *sni.ReadMemoryRequest) (*sni.ReadMemoryResponse, error) {
	panic("implement me")
}

func (s *devicesService) WriteMemory(ctx context.Context, request *sni.WriteMemoryRequest) (*sni.WriteMemoryResponse, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *devicesService) AcquireDevice(uri string) (dev snes.Queue, err error) {

	var ok bool
	s.devicesRw.RLock()
	dev, ok = s.devices[uri]
	s.devicesRw.RUnlock()
	if ok {
		return
	}

	var u *url.URL
	u, err = url.Parse(uri)
	if err != nil {
		return
	}

	var drv snes.Driver
	drv, ok = snes.DriverByName(u.Scheme)
	if !ok {
		err = fmt.Errorf("driver not found by name '%s'", u.Scheme)
		return
	}

	desc := drv.Empty()
	desc.Base().Id = u.Opaque
	dev, err = drv.Open(desc)
	if err != nil {
		return
	}

	s.devicesRw.Lock()
	s.devices[uri] = dev
	s.devicesRw.Unlock()
	return
}
