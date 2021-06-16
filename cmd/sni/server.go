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
		if !kindPredicate(driver.Driver.Kind()) {
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

	gerr = snes.UseDeviceMemory(rctx, uri, func(mctx context.Context, memory snes.DeviceMemory) (err error) {
		var mrsp []snes.MemoryReadResponse
		mrsp, err = memory.MultiReadMemory(mctx, snes.MemoryReadRequest{
			Address: request.Request.Address,
			Size:    int(request.Request.Size),
		})
		if err != nil {
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

	gerr = snes.UseDeviceMemory(rctx, uri, func(mctx context.Context, memory snes.DeviceMemory) (err error) {
		var mrsp []snes.MemoryWriteResponse
		mrsp, err = memory.MultiWriteMemory(mctx, snes.MemoryWriteRequest{
			Address: request.Request.Address,
			Data:    request.Request.Data,
		})
		if err != nil {
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

	if gerr != nil {
		rsp = nil
		return
	}
	return
}

func (s *deviceMemoryService) MultiRead(
	gctx context.Context,
	request *sni.MultiReadMemoryRequest,
) (grsp *sni.MultiReadMemoryResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	var grsps []*sni.ReadMemoryResponse
	gerr = snes.UseDeviceMemory(gctx, uri, func(mctx context.Context, memory snes.DeviceMemory) (err error) {
		reads := make([]snes.MemoryReadRequest, 0, len(request.Requests))
		for _, req := range request.Requests {
			reads = append(reads, snes.MemoryReadRequest{
				Address: req.Address,
				Size:    int(req.Size),
			})
		}

		var mrsps []snes.MemoryReadResponse
		mrsps, err = memory.MultiReadMemory(mctx, reads...)
		if err != nil {
			return
		}

		grsps = make([]*sni.ReadMemoryResponse, 0, len(mrsps))
		for _, mrsp := range mrsps {
			grsps = append(grsps, &sni.ReadMemoryResponse{
				Address: mrsp.Address,
				Data:    mrsp.Data,
			})
		}
		return
	})
	if gerr != nil {
		grsp = nil
		return
	}

	grsp = &sni.MultiReadMemoryResponse{
		Uri:       request.Uri,
		Responses: grsps,
	}
	return
}

func (s *deviceMemoryService) MultiWrite(
	gctx context.Context,
	request *sni.MultiWriteMemoryRequest,
) (grsp *sni.MultiWriteMemoryResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	var grsps []*sni.WriteMemoryResponse
	gerr = snes.UseDeviceMemory(gctx, uri, func(mctx context.Context, memory snes.DeviceMemory) (err error) {
		writes := make([]snes.MemoryWriteRequest, 0, len(request.Requests))
		for _, req := range request.Requests {
			writes = append(writes, snes.MemoryWriteRequest{
				Address: req.Address,
				Data:    req.Data,
			})
		}

		var mrsps []snes.MemoryWriteResponse
		mrsps, err = memory.MultiWriteMemory(mctx, writes...)
		if err != nil {
			return
		}

		grsps = make([]*sni.WriteMemoryResponse, 0, len(mrsps))
		for _, mrsp := range mrsps {
			grsps = append(grsps, &sni.WriteMemoryResponse{
				Address: mrsp.Address,
				Size:    uint32(mrsp.Size),
			})
		}
		return
	})
	if gerr != nil {
		grsp = nil
		return
	}

	grsp = &sni.MultiWriteMemoryResponse{
		Uri:       request.Uri,
		Responses: grsps,
	}
	return
}
