package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
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
}

func makeBool(v bool) *bool {
	return &v
}

func (s *memoryUnaryService) ReadMemory(rctx context.Context, request *sni.ReadMemoryRequest) (rsp *sni.ReadMemoryResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	gerr = snes.UseDevice(rctx, uri, func(ctx context.Context, dev snes.Device) (err error) {
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
