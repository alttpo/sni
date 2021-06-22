package main

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
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

func (s *deviceMemoryService) MappingDetect(gctx context.Context, request *sni.DetectMemoryMappingRequest) (grsp *sni.DetectMemoryMappingResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}
	if request.RomHeader00FFB0 != nil && len(request.RomHeader00FFB0) < 0x30 {
		gerr = status.Error(codes.InvalidArgument, "input ROM header must be at least $30 bytes")
		return
	}

	gerr = snes.UseDeviceMemory(
		gctx,
		uri,
		// TODO: this capability is optional; validate it in the call itself when it is about to be used:
		[]sni.DeviceCapability{sni.DeviceCapability_ReadMemory},
		func(mctx context.Context, memory snes.DeviceMemory) (err error) {
			var memoryMapping sni.MemoryMapping
			var confidence bool
			var outHeaderBytes []byte
			memoryMapping, confidence, outHeaderBytes, err = mapping.Detect(
				mctx,
				memory,
				request.FallbackMemoryMapping,
				request.RomHeader00FFB0,
			)

			grsp = &sni.DetectMemoryMappingResponse{
				Uri:             request.GetUri(),
				MemoryMapping:   memoryMapping,
				Confidence:      confidence,
				RomHeader00FFB0: outHeaderBytes,
			}
			return
		},
	)

	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
		return
	}
	return
}

func (s *deviceMemoryService) SingleRead(
	rctx context.Context,
	request *sni.SingleReadMemoryRequest,
) (grsp *sni.SingleReadMemoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	gerr = snes.UseDeviceMemory(
		rctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_ReadMemory},
		func(mctx context.Context, memory snes.DeviceMemory) (err error) {
			var mrsp []snes.MemoryReadResponse
			mrsp, err = memory.MultiReadMemory(mctx, snes.MemoryReadRequest{
				RequestAddress:      request.Request.GetRequestAddress(),
				RequestAddressSpace: request.Request.GetRequestAddressSpace(),
				RequestMapping:      request.Request.GetRequestMemoryMapping(),
				Size:                int(request.Request.GetSize()),
			})
			if err != nil {
				return
			}
			if len(mrsp) != 1 {
				err = status.Error(codes.Internal, "internal bug: single read must have a single response")
				return
			}
			if actual, expected := uint32(len(mrsp[0].Data)), request.Request.GetSize(); actual != expected {
				err = status.Errorf(
					codes.Internal,
					"internal bug: single read must return data of the requested size; actual $%x expected $%x",
					actual,
					expected,
				)
				return
			}

			grsp = &sni.SingleReadMemoryResponse{
				Uri: request.Uri,
				Response: &sni.ReadMemoryResponse{
					RequestAddress:       mrsp[0].RequestAddress,
					RequestAddressSpace:  mrsp[0].RequestAddressSpace,
					RequestMemoryMapping: mrsp[0].RequestMapping,
					DeviceAddress:        mrsp[0].DeviceAddress,
					DeviceAddressSpace:   mrsp[0].DeviceAddressSpace,
					Data:                 mrsp[0].Data,
				},
			}
			return
		},
	)

	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
		return
	}
	return
}

func (s *deviceMemoryService) SingleWrite(
	rctx context.Context,
	request *sni.SingleWriteMemoryRequest,
) (grsp *sni.SingleWriteMemoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	gerr = snes.UseDeviceMemory(
		rctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_WriteMemory},
		func(mctx context.Context, memory snes.DeviceMemory) (err error) {
			var mrsp []snes.MemoryWriteResponse
			mrsp, err = memory.MultiWriteMemory(mctx, snes.MemoryWriteRequest{
				RequestAddress:      request.Request.GetRequestAddress(),
				RequestAddressSpace: request.Request.GetRequestAddressSpace(),
				RequestMapping:      request.Request.GetRequestMemoryMapping(),
				Data:                request.Request.GetData(),
			})
			if err != nil {
				return
			}
			if len(mrsp) != 1 {
				err = status.Error(codes.Internal, "internal bug: single write must have a single response")
				return
			}
			if actual, expected := mrsp[0].Size, len(request.Request.GetData()); actual != expected {
				err = status.Errorf(
					codes.Internal,
					"internal bug: single write must return size of the written data; actual $%x expected $%x",
					actual,
					expected,
				)
				return
			}

			grsp = &sni.SingleWriteMemoryResponse{
				Uri: request.Uri,
				Response: &sni.WriteMemoryResponse{
					RequestAddress:       mrsp[0].RequestAddress,
					RequestAddressSpace:  mrsp[0].RequestAddressSpace,
					RequestMemoryMapping: mrsp[0].RequestMapping,
					DeviceAddress:        mrsp[0].DeviceAddress,
					DeviceAddressSpace:   mrsp[0].DeviceAddressSpace,
					Size:                 uint32(mrsp[0].Size),
				},
			}
			return
		},
	)

	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
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
	gerr = snes.UseDeviceMemory(
		gctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_ReadMemory},
		func(mctx context.Context, memory snes.DeviceMemory) (err error) {
			reads := make([]snes.MemoryReadRequest, 0, len(request.Requests))
			for _, req := range request.Requests {
				reads = append(reads, snes.MemoryReadRequest{
					RequestAddress:      req.GetRequestAddress(),
					RequestAddressSpace: req.GetRequestAddressSpace(),
					RequestMapping:      req.GetRequestMemoryMapping(),
					Size:                int(req.GetSize()),
				})
			}

			var mrsps []snes.MemoryReadResponse
			mrsps, err = memory.MultiReadMemory(mctx, reads...)
			if err != nil {
				return
			}
			if actual, expected := len(mrsps), len(reads); actual != expected {
				err = status.Errorf(
					codes.Internal,
					"internal bug: multi read must have equal number of responses and requests; actual %d expected %d",
					actual,
					expected,
				)
				return
			}

			grsps = make([]*sni.ReadMemoryResponse, 0, len(mrsps))
			for j, mrsp := range mrsps {
				if actual, expected := len(mrsp.Data), reads[j].Size; actual != expected {
					err = status.Errorf(
						codes.Internal,
						"internal bug: read[%d] must return data of the requested size; actual $%x expected $%x",
						j,
						actual,
						expected,
					)
					return
				}

				grsps = append(grsps, &sni.ReadMemoryResponse{
					RequestAddress:       mrsp.RequestAddress,
					RequestAddressSpace:  mrsp.RequestAddressSpace,
					RequestMemoryMapping: mrsp.RequestMapping,
					DeviceAddress:        mrsp.DeviceAddress,
					DeviceAddressSpace:   mrsp.DeviceAddressSpace,
					Data:                 mrsp.Data,
				})
			}
			return
		},
	)
	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
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
	gerr = snes.UseDeviceMemory(
		gctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_WriteMemory},
		func(mctx context.Context, memory snes.DeviceMemory) (err error) {
			writes := make([]snes.MemoryWriteRequest, 0, len(request.Requests))
			for _, req := range request.Requests {
				writes = append(writes, snes.MemoryWriteRequest{
					RequestAddress:      req.GetRequestAddress(),
					RequestAddressSpace: req.GetRequestAddressSpace(),
					RequestMapping:      req.GetRequestMemoryMapping(),
					Data:                req.Data,
				})
			}

			var mrsps []snes.MemoryWriteResponse
			mrsps, err = memory.MultiWriteMemory(mctx, writes...)
			if err != nil {
				return
			}
			if actual, expected := len(mrsps), len(writes); actual != expected {
				err = status.Errorf(
					codes.Internal,
					"internal bug: multi write must have equal number of responses and requests; actual %d expected %d",
					actual,
					expected,
				)
				return
			}

			grsps = make([]*sni.WriteMemoryResponse, 0, len(mrsps))
			for j, mrsp := range mrsps {
				if actual, expected := mrsp.Size, len(writes[j].Data); actual != expected {
					err = status.Errorf(
						codes.Internal,
						"internal bug: write[%d] must return size of the written data; actual $%x expected $%x",
						j,
						actual,
						expected,
					)
					return
				}

				grsps = append(grsps, &sni.WriteMemoryResponse{
					RequestAddress:       mrsp.RequestAddress,
					RequestAddressSpace:  mrsp.RequestAddressSpace,
					RequestMemoryMapping: mrsp.RequestMapping,
					DeviceAddress:        mrsp.DeviceAddress,
					DeviceAddressSpace:   mrsp.DeviceAddressSpace,
					Size:                 uint32(mrsp.Size),
				})
			}
			return
		},
	)
	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
		return
	}

	grsp = &sni.MultiWriteMemoryResponse{
		Uri:       request.Uri,
		Responses: grsps,
	}
	return
}

type deviceControlService struct {
	sni.UnimplementedDeviceControlServer
}

func (d *deviceControlService) ResetSystem(gctx context.Context, request *sni.ResetSystemRequest) (grsp *sni.ResetSystemResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	gerr = snes.UseDeviceControl(
		gctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_ResetSystem},
		func(mctx context.Context, control snes.DeviceControl) (err error) {
			return control.ResetSystem(mctx)
		},
	)
	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
		return
	}

	grsp = &sni.ResetSystemResponse{
		Uri: request.Uri,
	}
	return
}

func (d *deviceControlService) PauseUnpauseEmulation(gctx context.Context, request *sni.PauseEmulationRequest) (grsp *sni.PauseEmulationResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	paused := false
	gerr = snes.UseDeviceControl(
		gctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_PauseUnpauseEmulation},
		func(mctx context.Context, control snes.DeviceControl) (err error) {
			paused, err = control.PauseUnpause(mctx, request.Paused)
			return
		},
	)
	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
		return
	}

	grsp = &sni.PauseEmulationResponse{
		Uri:    request.Uri,
		Paused: paused,
	}
	return
}

func (d *deviceControlService) PauseToggleEmulation(gctx context.Context, request *sni.PauseToggleEmulationRequest) (grsp *sni.PauseToggleEmulationResponse, gerr error) {
	uri, err := url.Parse(request.Uri)
	if err != nil {
		gerr = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	gerr = snes.UseDeviceControl(
		gctx,
		uri,
		[]sni.DeviceCapability{sni.DeviceCapability_PauseToggleEmulation},
		func(mctx context.Context, control snes.DeviceControl) (err error) {
			return control.PauseToggle(mctx)
		},
	)
	if gerr != nil {
		var coded *snes.CodedError
		if errors.As(gerr, &coded) {
			gerr = status.Error(coded.Code, coded.Error())
		} else {
			grsp = nil
		}
		return
	}

	grsp = &sni.PauseToggleEmulationResponse{
		Uri: request.Uri,
	}
	return
}
