package grpcimpl

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/devices"
	"sni/protos/sni"
)

type DeviceInfoService struct {
	sni.UnimplementedDeviceInfoServer
}

func (d *DeviceInfoService) FetchFields(gctx context.Context, request *sni.FieldsRequest) (grsp *sni.FieldsResponse, gerr error) {
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

	if _, err := driver.HasCapabilities(sni.DeviceCapability_FetchFields); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var values []string
	values, gerr = device.FetchFields(gctx, request.Fields...)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	if len(values) != len(request.Fields) {
		return nil, status.Error(codes.Internal, "values slice length did not match fields slice length")
	}

	grsp = &sni.FieldsResponse{
		Uri:    request.Uri,
		Fields: request.Fields,
		Values: values,
	}

	return
}
