package grpcimpl

import (
	"bytes"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/devices"
	"sni/protos/sni"
)

type DeviceFilesystem struct {
	sni.UnimplementedDeviceFilesystemServer
}

func (d *DeviceFilesystem) ReadDirectory(ctx context.Context, request *sni.ReadDirectoryRequest) (grsp *sni.ReadDirectoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if _, err := driver.HasCapabilities(sni.DeviceCapability_ReadDirectory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var files []devices.DirEntry
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

func (d *DeviceFilesystem) MakeDirectory(ctx context.Context, request *sni.MakeDirectoryRequest) (grsp *sni.MakeDirectoryResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
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

func (d *DeviceFilesystem) RemoveFile(ctx context.Context, request *sni.RemoveFileRequest) (grsp *sni.RemoveFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
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

func (d *DeviceFilesystem) RenameFile(ctx context.Context, request *sni.RenameFileRequest) (grsp *sni.RenameFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
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

func (d *DeviceFilesystem) PutFile(ctx context.Context, request *sni.PutFileRequest) (grsp *sni.PutFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
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

	// translate response:
	grsp = &sni.PutFileResponse{
		Uri:  request.Uri,
		Path: request.Path,
		Size: n,
	}
	return
}

func (d *DeviceFilesystem) GetFile(ctx context.Context, request *sni.GetFileRequest) (grsp *sni.GetFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
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

func (d *DeviceFilesystem) BootFile(ctx context.Context, request *sni.BootFileRequest) (grsp *sni.BootFileResponse, gerr error) {
	uri, err := url.Parse(request.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var driver devices.Driver
	var device devices.AutoCloseableDevice
	driver, device, gerr = devices.DeviceByUri(uri)
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
