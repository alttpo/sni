package main

import (
	"bytes"
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
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

func grpcError(err error) error {
	var coded *snes.CodedError
	if errors.As(err, &coded) {
		return status.Error(coded.Code, coded.Error())
	}
	return err
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

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	// require ReadMemory capability if ROM header data is not provided:
	if request.RomHeader00FFB0 == nil {
		if _, err := driver.HasCapabilities(sni.DeviceCapability_ReadMemory); err != nil {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}
	}

	var memoryMapping sni.MemoryMapping
	var confidence bool
	var outHeaderBytes []byte
	memoryMapping, confidence, outHeaderBytes, gerr = mapping.Detect(
		gctx,
		device,
		request.FallbackMemoryMapping,
		request.RomHeader00FFB0,
	)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	grsp = &sni.DetectMemoryMappingResponse{
		Uri:             request.GetUri(),
		MemoryMapping:   memoryMapping,
		Confidence:      confidence,
		RomHeader00FFB0: outHeaderBytes,
	}

	return
}

func (s *deviceMemoryService) SingleRead(
	gctx context.Context,
	request *sni.SingleReadMemoryRequest,
) (grsp *sni.SingleReadMemoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_ReadMemory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var mrsp []snes.MemoryReadResponse
	mrsp, gerr = device.MultiReadMemory(gctx, snes.MemoryReadRequest{
		RequestAddress: snes.AddressTuple{
			Address:       request.Request.GetRequestAddress(),
			AddressSpace:  request.Request.GetRequestAddressSpace(),
			MemoryMapping: request.Request.GetRequestMemoryMapping(),
		},
		Size: int(request.Request.GetSize()),
	})
	if gerr != nil {
		return nil, grpcError(gerr)
	}
	if len(mrsp) != 1 {
		err = status.Error(codes.Internal, "single read must have a single response")
		return
	}
	if actual, expected := uint32(len(mrsp[0].Data)), request.Request.GetSize(); actual != expected {
		gerr = status.Errorf(
			codes.Internal,
			"single read must return data of the requested size; actual $%x expected $%x",
			actual,
			expected,
		)
		return
	}

	grsp = &sni.SingleReadMemoryResponse{
		Uri: request.Uri,
		Response: &sni.ReadMemoryResponse{
			RequestAddress:       mrsp[0].RequestAddress.Address,
			RequestAddressSpace:  mrsp[0].RequestAddress.AddressSpace,
			RequestMemoryMapping: mrsp[0].RequestAddress.MemoryMapping,
			DeviceAddress:        mrsp[0].DeviceAddress.Address,
			DeviceAddressSpace:   mrsp[0].DeviceAddress.AddressSpace,
			Data:                 mrsp[0].Data,
		},
	}

	return
}

func (s *deviceMemoryService) SingleWrite(
	gctx context.Context,
	request *sni.SingleWriteMemoryRequest,
) (grsp *sni.SingleWriteMemoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_WriteMemory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var mrsp []snes.MemoryWriteResponse
	mrsp, gerr = device.MultiWriteMemory(gctx, snes.MemoryWriteRequest{
		RequestAddress: snes.AddressTuple{
			Address:       request.Request.GetRequestAddress(),
			AddressSpace:  request.Request.GetRequestAddressSpace(),
			MemoryMapping: request.Request.GetRequestMemoryMapping(),
		},
		Data: request.Request.GetData(),
	})
	if gerr != nil {
		return nil, grpcError(gerr)
	}
	if len(mrsp) != 1 {
		return nil, status.Error(codes.Internal, "single write must have a single response")
	}
	if actual, expected := mrsp[0].Size, len(request.Request.GetData()); actual != expected {
		gerr = status.Errorf(
			codes.Internal,
			"single write must return size of the written data; actual $%x expected $%x",
			actual,
			expected,
		)
		return
	}

	grsp = &sni.SingleWriteMemoryResponse{
		Uri: request.Uri,
		Response: &sni.WriteMemoryResponse{
			RequestAddress:       mrsp[0].RequestAddress.Address,
			RequestAddressSpace:  mrsp[0].RequestAddress.AddressSpace,
			RequestMemoryMapping: mrsp[0].RequestAddress.MemoryMapping,
			DeviceAddress:        mrsp[0].DeviceAddress.Address,
			DeviceAddressSpace:   mrsp[0].DeviceAddress.AddressSpace,
			Size:                 uint32(mrsp[0].Size),
		},
	}

	return
}

func (s *deviceMemoryService) MultiRead(
	gctx context.Context,
	request *sni.MultiReadMemoryRequest,
) (grsp *sni.MultiReadMemoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_ReadMemory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var grsps []*sni.ReadMemoryResponse
	reads := make([]snes.MemoryReadRequest, 0, len(request.Requests))
	for _, req := range request.Requests {
		reads = append(reads, snes.MemoryReadRequest{
			RequestAddress: snes.AddressTuple{
				Address:       req.GetRequestAddress(),
				AddressSpace:  req.GetRequestAddressSpace(),
				MemoryMapping: req.GetRequestMemoryMapping(),
			},
			Size: int(req.GetSize()),
		})
	}

	var mrsps []snes.MemoryReadResponse
	mrsps, gerr = device.MultiReadMemory(gctx, reads...)
	if gerr != nil {
		return nil, grpcError(gerr)
	}
	if actual, expected := len(mrsps), len(reads); actual != expected {
		gerr = status.Errorf(
			codes.Internal,
			"multi read must have equal number of responses and requests; actual %d expected %d",
			actual,
			expected,
		)
		return
	}

	grsps = make([]*sni.ReadMemoryResponse, 0, len(mrsps))
	for j, mrsp := range mrsps {
		if actual, expected := len(mrsp.Data), reads[j].Size; actual != expected {
			gerr = status.Errorf(
				codes.Internal,
				"read[%d] must return data of the requested size; actual $%x expected $%x",
				j,
				actual,
				expected,
			)
			return
		}

		grsps = append(grsps, &sni.ReadMemoryResponse{
			RequestAddress:       mrsp.RequestAddress.Address,
			RequestAddressSpace:  mrsp.RequestAddress.AddressSpace,
			RequestMemoryMapping: mrsp.RequestAddress.MemoryMapping,
			DeviceAddress:        mrsp.DeviceAddress.Address,
			DeviceAddressSpace:   mrsp.DeviceAddress.AddressSpace,
			Data:                 mrsp.Data,
		})
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
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_WriteMemory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	writes := make([]snes.MemoryWriteRequest, 0, len(request.Requests))
	for _, req := range request.Requests {
		writes = append(writes, snes.MemoryWriteRequest{
			RequestAddress: snes.AddressTuple{
				Address:       req.GetRequestAddress(),
				AddressSpace:  req.GetRequestAddressSpace(),
				MemoryMapping: req.GetRequestMemoryMapping(),
			},
			Data: req.Data,
		})
	}

	var mrsps []snes.MemoryWriteResponse
	mrsps, gerr = device.MultiWriteMemory(gctx, writes...)
	if gerr != nil {
		return nil, grpcError(gerr)
	}
	if actual, expected := len(mrsps), len(writes); actual != expected {
		gerr = status.Errorf(
			codes.Internal,
			"multi write must have equal number of responses and requests; actual %d expected %d",
			actual,
			expected,
		)
		return
	}

	grsps := make([]*sni.WriteMemoryResponse, 0, len(mrsps))
	for j, mrsp := range mrsps {
		if actual, expected := mrsp.Size, len(writes[j].Data); actual != expected {
			gerr = status.Errorf(
				codes.Internal,
				"write[%d] must return size of the written data; actual $%x expected $%x",
				j,
				actual,
				expected,
			)
			return
		}

		grsps = append(grsps, &sni.WriteMemoryResponse{
			RequestAddress:       mrsp.RequestAddress.Address,
			RequestAddressSpace:  mrsp.RequestAddress.AddressSpace,
			RequestMemoryMapping: mrsp.RequestAddress.MemoryMapping,
			DeviceAddress:        mrsp.DeviceAddress.Address,
			DeviceAddressSpace:   mrsp.DeviceAddress.AddressSpace,
			Size:                 uint32(mrsp.Size),
		})
	}

	grsp = &sni.MultiWriteMemoryResponse{
		Uri:       request.Uri,
		Responses: grsps,
	}
	return
}

func (s *deviceMemoryService) StreamRead(stream sni.DeviceMemory_StreamReadServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		grsp, gerr := s.MultiRead(stream.Context(), in)
		if gerr != nil {
			// TODO: stream errors as responses?
			return gerr
		}
		err = stream.Send(grsp)
		if err != nil {
			return err
		}
	}
}

func (s *deviceMemoryService) StreamWrite(stream sni.DeviceMemory_StreamWriteServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		grsp, gerr := s.MultiWrite(stream.Context(), in)
		if gerr != nil {
			// TODO: stream errors as responses?
			return gerr
		}
		err = stream.Send(grsp)
		if err != nil {
			return err
		}
	}
}

type deviceControlService struct {
	sni.UnimplementedDeviceControlServer
}

func (d *deviceControlService) ResetSystem(gctx context.Context, request *sni.ResetSystemRequest) (grsp *sni.ResetSystemResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_ResetSystem); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	gerr = device.ResetSystem(gctx)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	grsp = &sni.ResetSystemResponse{
		Uri: request.Uri,
	}

	return
}

func (d *deviceControlService) PauseUnpauseEmulation(gctx context.Context, request *sni.PauseEmulationRequest) (grsp *sni.PauseEmulationResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_PauseUnpauseEmulation); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var paused bool
	paused, gerr = device.PauseUnpause(gctx, request.Paused)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	grsp = &sni.PauseEmulationResponse{
		Uri:    request.Uri,
		Paused: paused,
	}

	return
}

func (d *deviceControlService) PauseToggleEmulation(gctx context.Context, request *sni.PauseToggleEmulationRequest) (grsp *sni.PauseToggleEmulationResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_PauseToggleEmulation); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	gerr = device.PauseToggle(gctx)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	grsp = &sni.PauseToggleEmulationResponse{
		Uri: request.Uri,
	}

	return
}

type deviceFilesystem struct {
	sni.UnimplementedDeviceFilesystemServer
}

func (d *deviceFilesystem) ReadDirectory(ctx context.Context, request *sni.ReadDirectoryRequest) (grsp *sni.ReadDirectoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_ReadDirectory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var files []snes.DirEntry
	files, gerr = device.ReadDirectory(ctx, request.GetPath())
	if gerr != nil {
		return
	}

	// translate response:
	grsp = &sni.ReadDirectoryResponse{
		Uri:     request.Uri,
		Path:    request.Path,
		Entries: make([]*sni.DirEntry, len(files)),
	}
	for i, file := range files {
		grsp.Entries[i] = &sni.DirEntry{
			Name: file.Name,
			Type: file.Type,
		}
	}
	return
}

func (d *deviceFilesystem) MakeDirectory(ctx context.Context, request *sni.MakeDirectoryRequest) (grsp *sni.MakeDirectoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_MakeDirectory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	gerr = device.MakeDirectory(ctx, request.GetPath())
	if gerr != nil {
		return
	}

	// translate response:
	grsp = &sni.MakeDirectoryResponse{
		Uri:  request.Uri,
		Path: request.Path,
	}
	return
}

func (d *deviceFilesystem) RemoveFile(ctx context.Context, request *sni.RemoveFileRequest) (grsp *sni.RemoveFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_RemoveFile); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	gerr = device.RemoveFile(ctx, request.GetPath())
	if gerr != nil {
		return
	}

	// translate response:
	grsp = &sni.RemoveFileResponse{
		Uri:  request.Uri,
		Path: request.Path,
	}
	return
}

func (d *deviceFilesystem) RenameFile(ctx context.Context, request *sni.RenameFileRequest) (grsp *sni.RenameFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_RenameFile); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	gerr = device.RenameFile(ctx, request.GetPath(), request.GetNewFilename())
	if gerr != nil {
		return
	}

	// translate response:
	grsp = &sni.RenameFileResponse{
		Uri:         request.Uri,
		Path:        request.Path,
		NewFilename: request.NewFilename,
	}
	return
}

func (d *deviceFilesystem) PutFile(ctx context.Context, request *sni.PutFileRequest) (grsp *sni.PutFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_PutFile); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var n uint32
	n, gerr = device.PutFile(
		ctx,
		request.GetPath(),
		uint32(len(request.GetData())),
		bytes.NewReader(request.GetData()),
		nil,
	)
	if gerr != nil {
		return
	}
	_ = n

	// translate response:
	grsp = &sni.PutFileResponse{
		Uri:  request.Uri,
		Path: request.Path,
	}
	return
}

func (d *deviceFilesystem) GetFile(ctx context.Context, request *sni.GetFileRequest) (grsp *sni.GetFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_GetFile); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	data := bytes.Buffer{}
	var n uint32
	n, gerr = device.GetFile(ctx, request.GetPath(), &data, nil, func(current uint32, total uint32) {
		// grow the buffer if we haven't already:
		if uint32(data.Cap()) < total {
			data.Grow(int(total - uint32(data.Cap())))
		}
	})
	if gerr != nil {
		return
	}

	// translate response:
	grsp = &sni.GetFileResponse{
		Uri:  request.Uri,
		Path: request.Path,
		Size: n,
		Data: data.Bytes(),
	}
	return
}

func (d *deviceFilesystem) BootFile(ctx context.Context, request *sni.BootFileRequest) (grsp *sni.BootFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver snes.Driver
	var device snes.AutoCloseableDevice
	driver, device, gerr = snes.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_BootFile); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	gerr = device.BootFile(ctx, request.GetPath())
	if gerr != nil {
		return
	}

	// translate response:
	grsp = &sni.BootFileResponse{
		Uri:  request.Uri,
		Path: request.Path,
	}
	return
}
