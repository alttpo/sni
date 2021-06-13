package main

import (
	"context"
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
				Uri:          descriptor.Uri.String(),
				DisplayName:  descriptor.DisplayName,
				Kind:         descriptor.Kind,
				Capabilities: int32(descriptor.Capabilities),
			})
		}
	}

	return &sni.DevicesResponse{Devices: devices}, nil
}

type deviceMemoryService struct {
	sni.UnimplementedDeviceMemoryServer
}

func makeBool(v bool) *bool {
	return &v
}

func (s *deviceMemoryService) Read(
	rctx context.Context,
	request *sni.SingleReadMemoryRequest,
) (rsp *sni.SingleReadMemoryResponse, gerr error) {
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
				Address: request.Request.Address,
				Size:    int(request.Request.Size),
			})
			if merr != nil {
				return
			}

			rsp = &sni.SingleReadMemoryResponse{
				Uri: request.Uri,
				Response: &sni.ReadMemoryResponse{
					Address: request.Request.Address,
					Data:    mrsp[0].Data,
				},
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

func (s *deviceMemoryService) Write(
	rctx context.Context,
	request *sni.SingleWriteMemoryRequest,
) (rsp *sni.SingleWriteMemoryResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	gerr = snes.UseDevice(rctx, uri, func(ctx context.Context, dev snes.Device) (err error) {
		// TODO: could offer stateful binding of device to peer
		//peer.FromContext(ctx)
		err = dev.UseMemory(ctx, func(mctx context.Context, memory snes.DeviceMemory) (merr error) {
			var mrsp []snes.MemoryWriteResponse
			mrsp, merr = memory.MultiWriteMemory(mctx, snes.MemoryWriteRequest{
				Address: request.Request.Address,
				Data:    request.Request.Data,
			})
			if merr != nil {
				return
			}

			rsp = &sni.SingleWriteMemoryResponse{
				Uri: request.Uri,
				Response: &sni.WriteMemoryResponse{
					Address: mrsp[0].Address,
					Size:    uint32(mrsp[0].Size),
				},
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
