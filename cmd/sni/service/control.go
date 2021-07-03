package service

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
)

type DeviceControlService struct {
	sni.UnimplementedDeviceControlServer
}

func (d *DeviceControlService) ResetSystem(gctx context.Context, request *sni.ResetSystemRequest) (grsp *sni.ResetSystemResponse, gerr error) {
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

func (d *DeviceControlService) PauseUnpauseEmulation(gctx context.Context, request *sni.PauseEmulationRequest) (grsp *sni.PauseEmulationResponse, gerr error) {
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

func (d *DeviceControlService) PauseToggleEmulation(gctx context.Context, request *sni.PauseToggleEmulationRequest) (grsp *sni.PauseToggleEmulationResponse, gerr error) {
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
