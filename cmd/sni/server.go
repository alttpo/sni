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
				Uri:                 descriptor.Uri.String(),
				DisplayName:         descriptor.DisplayName,
				Kind:                descriptor.Kind,
				Capabilities:        descriptor.Capabilities,
				DefaultAddressSpace: descriptor.DefaultAddressSpace,
			})
		}
	}

	return &sni.DevicesResponse{Devices: devices}, nil
}

type deviceMemoryService struct {
	sni.UnimplementedDeviceMemoryServer
}

//func (s *deviceMemoryService) MappingDetect(context.Context, *sni.DetectMemoryMappingRequest) (*sni.MemoryMappingResponse, error) {
//	return nil, status.Errorf(codes.Unimplemented, "method MappingDetect not implemented")
//}
//func (s *deviceMemoryService) MappingSet(context.Context, *sni.SetMemoryMappingRequest) (*sni.MemoryMappingResponse, error) {
//	return nil, status.Errorf(codes.Unimplemented, "method MappingSet not implemented")
//}
//func (s *deviceMemoryService) MappingGet(context.Context, *sni.GetMemoryMappingRequest) (*sni.MemoryMappingResponse, error) {
//	return nil, status.Errorf(codes.Unimplemented, "method MappingGet not implemented")
//}

func (s *deviceMemoryService) SingleRead(
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
			RequestAddress: request.Request.RequestAddress,
			Size:           int(request.Request.Size),
		})
		if err != nil {
			return
		}

		rsp = &sni.SingleReadMemoryResponse{
			Uri: request.Uri,
			Response: &sni.ReadMemoryResponse{
				RequestAddress:      mrsp[0].RequestAddress,
				RequestAddressSpace: mrsp[0].RequestAddressSpace,
				DeviceAddress:       mrsp[0].DeviceAddress,
				DeviceAddressSpace:  mrsp[0].DeviceAddressSpace,
				Data:                mrsp[0].Data,
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

func (s *deviceMemoryService) SingleWrite(
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
			RequestAddress:      request.Request.GetRequestAddress(),
			RequestAddressSpace: request.Request.GetRequestAddressSpace(),
			Data:                request.Request.GetData(),
		})
		if err != nil {
			return
		}

		rsp = &sni.SingleWriteMemoryResponse{
			Uri: request.Uri,
			Response: &sni.WriteMemoryResponse{
				RequestAddress:      mrsp[0].RequestAddress,
				RequestAddressSpace: mrsp[0].RequestAddressSpace,
				DeviceAddress:       mrsp[0].DeviceAddress,
				DeviceAddressSpace:  mrsp[0].DeviceAddressSpace,
				Size:                uint32(mrsp[0].Size),
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
				RequestAddress:      req.GetRequestAddress(),
				RequestAddressSpace: req.GetRequestAddressSpace(),
				Size:                int(req.GetSize()),
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
				RequestAddress:      mrsp.RequestAddress,
				RequestAddressSpace: mrsp.RequestAddressSpace,
				DeviceAddress:       mrsp.DeviceAddress,
				DeviceAddressSpace:  mrsp.DeviceAddressSpace,
				Data:                mrsp.Data,
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
				RequestAddress:      req.GetRequestAddress(),
				RequestAddressSpace: req.GetRequestAddressSpace(),
				Data:                req.Data,
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
				RequestAddress:      mrsp.RequestAddress,
				RequestAddressSpace: mrsp.RequestAddressSpace,
				DeviceAddress:       mrsp.DeviceAddress,
				DeviceAddressSpace:  mrsp.DeviceAddressSpace,
				Size:                uint32(mrsp.Size),
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
