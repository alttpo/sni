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
				Uri:         fmt.Sprintf("%s:%s", driver.Name, descriptor.GetId()),
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

func makeBool(v bool) *bool {
	return &v
}

func (s *devicesService) ReadMemory(ctx context.Context, request *sni.ReadMemoryRequest) (rsp *sni.ReadMemoryResponse, err error) {
	var dev snes.Queue
	dev, err = s.AcquireDevice(request.Uri)
	if err != nil {
		return
	}

	complete := make(chan struct{})

	addr := request.GetAddress()
	size := request.GetSize()
	reads := make([]snes.Read, 0, 8)
	data := make([]byte, 0, size)
	for size > 0 {
		chunkSize := uint32(255)
		if size < chunkSize {
			chunkSize = size
		}

		reads = append(reads, snes.Read{
			Address:    addr,
			Size:       uint8(chunkSize),
			Extra:      nil,
			Completion: func(response snes.Response) {
				data = append(data, response.Data...)
			},
		})

		size -= 255
		addr += 255
	}

	seq := dev.MakeReadCommands(reads, func(command snes.Command, cmderr error) {
		err = cmderr
		close(complete)
	})
	// enqueue the read:
	err = seq.EnqueueTo(dev)
	if err != nil {
		return
	}

	// wait until canceled or read completed:
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-complete:
		break
	}

	// err could be assigned from batchComplete callback:
	if err != nil {
		return
	}

	rsp = &sni.ReadMemoryResponse{
		Uri:     request.Uri,
		Success: makeBool(true),
		Error:   nil,
		Data:    data,
	}
	return
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
	if s.devices == nil {
		s.devices = make(map[string]snes.Queue)
	}
	s.devices[uri] = dev
	s.devicesRw.Unlock()
	return
}
