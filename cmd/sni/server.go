package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	devices   map[string]snes.Queue
}

func makeBool(v bool) *bool {
	return &v
}

func (s *memoryUnaryService) ReadMemory(ctx context.Context, request *sni.ReadMemoryRequest) (rsp *sni.ReadMemoryResponse, gerr error) {
	gerr = s.UseDevice(request.Uri, func(dev snes.Queue) (err error) {
		//peer.FromContext(ctx)

		addr := request.GetAddress()
		size := int32(request.GetSize())
		reads := make([]snes.Read, 0, 8)
		data := make([]byte, 0, size)
		for size > 0 {
			chunkSize := int32(255)
			if size < chunkSize {
				chunkSize = size
			}

			reads = append(reads, snes.Read{
				Address: addr,
				Size:    uint8(chunkSize),
				Extra:   nil,
				Completion: func(response snes.Response) {
					data = append(data, response.Data...)
				},
			})

			size -= 255
			addr += 255
		}

		wg := sync.WaitGroup{}
		var batchErrors chan error
		seq := dev.MakeReadCommands(reads, func(command snes.Command, cmderr error) {
			batchErrors <- cmderr
			wg.Done()
		})
		wg.Add(len(seq))
		batchErrors = make(chan error, len(seq))

		// enqueue the read:
		err = seq.EnqueueTo(dev)
		if err != nil {
			return
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		// wait until completed or canceled:
	waitLoop:
		for {
			select {
			case <-done:
				// read complete:
				break waitLoop
			case <-ctx.Done():
				// canceled:
				return ctx.Err()
			case berr := <-batchErrors:
				// take only the first error of the batches:
				if berr != nil && err == nil {
					err = status.Error(codes.Unavailable, berr.Error())
				}
				break
			}
		}

		rsp = &sni.ReadMemoryResponse{
			Uri:     request.Uri,
			Address: request.Address,
			Data:    data,
		}

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

func (s *memoryUnaryService) UseDevice(uri string, use func(snes.Queue) error) (err error) {
	var dev snes.Queue
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
	}

	err = use(dev)

	if dev.IsClosed() {
		s.devicesRw.Lock()
		if s.devices == nil {
			s.devices = make(map[string]snes.Queue)
		}
		delete(s.devices, uri)
		s.devicesRw.Unlock()
	}

	return
}
