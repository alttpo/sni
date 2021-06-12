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
	sni.UnimplementedDevicesServer
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

type memoryUnaryService struct {
	sni.UnimplementedMemoryUnaryServer

	// track opened devices by URI
	devicesRw sync.RWMutex
	devices   map[string]snes.Device
}

func makeBool(v bool) *bool {
	return &v
}

func (s *memoryUnaryService) ReadMemory(rctx context.Context, request *sni.ReadMemoryRequest) (rsp *sni.ReadMemoryResponse, gerr error) {
	gerr = s.UseDevice(rctx, request.Uri, func(ctx context.Context, dev snes.Device) (err error) {
		// TODO: could offer stateful binding of device to peer
		//peer.FromContext(ctx)
		err = dev.UseMemory(ctx, func(mctx context.Context, memory snes.DeviceMemory) (merr error) {
			var mrsp []snes.MemoryReadResponse
			mrsp, merr = memory.MultiReadMemory(mctx, snes.MemoryReadRequest{
				Address: request.Address,
				Size:    int(request.Size),
			})
			if merr != nil {
				return
			}

			rsp = &sni.ReadMemoryResponse{
				Uri:     request.Uri,
				Address: request.Address,
				Data:    mrsp[0].Data,
			}
			return
		})

		return
	})

	if gerr != nil {
		rsp = nil
		return
	}
	return
}

func (s *memoryUnaryService) WriteMemory(ctx context.Context, request *sni.WriteMemoryRequest) (*sni.WriteMemoryResponse, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *memoryUnaryService) UseDevice(ctx context.Context, uri string, use func(context.Context, snes.Device) error) (err error) {
	var dev snes.Device
	var ok bool

	s.devicesRw.RLock()
	dev, ok = s.devices[uri]
	s.devicesRw.RUnlock()

	if !ok {
		var u *url.URL
		u, err = url.Parse(uri)
		if err != nil {
			return
		}

		var gendrv snes.Driver
		gendrv, ok = snes.DriverByName(u.Scheme)
		if !ok {
			err = fmt.Errorf("driver not found by name '%s'", u.Scheme)
			return
		}
		drv, ok := gendrv.(snes.DeviceDriver)
		if !ok {
			err = fmt.Errorf("driver named '%s' is not a DeviceDriver", u.Scheme)
			return
		}

		desc := gendrv.Empty()
		desc.Base().Id = u.Opaque
		dev, err = drv.OpenDevice(desc)
		if err != nil {
			return
		}

		s.devicesRw.Lock()
		if s.devices == nil {
			s.devices = make(map[string]snes.Device)
		}
		s.devices[uri] = dev
		s.devicesRw.Unlock()
	}

	err = use(ctx, dev)

	if dev.IsClosed() {
		s.devicesRw.Lock()
		if s.devices == nil {
			s.devices = make(map[string]snes.Device)
		}
		delete(s.devices, uri)
		s.devicesRw.Unlock()
	}

	return
}
