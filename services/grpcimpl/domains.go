package grpcimpl

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/devices"
	"sni/protos/sni"
)

type DeviceMemoryDomainsService struct {
	sni.UnimplementedDeviceMemoryDomainsServer
}

func (s *DeviceMemoryDomainsService) MemoryDomains(ctx context.Context, request *sni.MemoryDomainsRequest) (grsp *sni.MemoryDomainsResponse, gerr error) {

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

	if _, err := driver.HasCapabilities(sni.DeviceCapability_ReadMemory); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

}

func (s *DeviceMemoryDomainsService) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (grsp *sni.MultiDomainReadResponse, gerr error) {
	//TODO implement me
	panic("implement me")
}

func (s *DeviceMemoryDomainsService) MultiDomainWrite(ctx context.Context, request *sni.MultiDomainWriteRequest) (grsp *sni.MultiDomainWriteResponse, gerr error) {
	//TODO implement me
	panic("implement me")
}

func (s *DeviceMemoryDomainsService) StreamDomainRead(server sni.DeviceMemoryDomains_StreamDomainReadServer) (gerr error) {
	//TODO implement me
	panic("implement me")
}

func (s *DeviceMemoryDomainsService) StreamDomainWrite(server sni.DeviceMemoryDomains_StreamDomainWriteServer) (gerr error) {
	//TODO implement me
	panic("implement me")
}
