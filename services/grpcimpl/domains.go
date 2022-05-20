package grpcimpl

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
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

	_ = driver

	gerr = device.EnsureOpened(func(device devices.Device) (err error) {
		md, ok := device.(devices.DeviceMemoryDomains)
		if !ok {
			return status.Error(codes.Unimplemented, "service not implemented for this driver")
		}

		grsp, err = md.MemoryDomains(ctx, request)
		return
	})
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	return
}

func (s *DeviceMemoryDomainsService) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (grsp *sni.MultiDomainReadResponse, gerr error) {
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

	gerr = device.EnsureOpened(func(device devices.Device) (err error) {
		md, ok := device.(devices.DeviceMemoryDomains)
		if !ok {
			return status.Error(codes.Unimplemented, "service not implemented for this driver")
		}

		grsp, err = md.MultiDomainRead(ctx, request)
		return
	})
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	return
}

func (s *DeviceMemoryDomainsService) MultiDomainWrite(ctx context.Context, request *sni.MultiDomainWriteRequest) (grsp *sni.MultiDomainWriteResponse, gerr error) {
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

	gerr = device.EnsureOpened(func(device devices.Device) (err error) {
		md, ok := device.(devices.DeviceMemoryDomains)
		if !ok {
			return status.Error(codes.Unimplemented, "service not implemented for this driver")
		}

		grsp, err = md.MultiDomainWrite(ctx, request)
		return
	})
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	return
}

func (s *DeviceMemoryDomainsService) StreamDomainRead(stream sni.DeviceMemoryDomains_StreamDomainReadServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		grsp, gerr := s.MultiDomainRead(stream.Context(), in)
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

func (s *DeviceMemoryDomainsService) StreamDomainWrite(stream sni.DeviceMemoryDomains_StreamDomainWriteServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		grsp, gerr := s.MultiDomainWrite(stream.Context(), in)
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
