package service

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/status"
	"sni/protos/sni"
	"sni/snes"
)

func grpcError(err error) error {
	var coded *snes.CodedError
	if errors.As(err, &coded) {
		return status.Error(coded.Code, coded.Error())
	}
	return err
}

type DevicesService struct {
	sni.UnimplementedDevicesServer
}

func (s *DevicesService) ListDevices(ctx context.Context, request *sni.DevicesRequest) (*sni.DevicesResponse, error) {
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

func (s *DevicesService) MethodRequestString(method string, req interface{}) string {
	if req == nil {
		return "nil"
	}

	return fmt.Sprintf("%+v", req)
}

func (s *DevicesService) MethodResponseString(method string, rsp interface{}) string {
	if rsp == nil {
		return "nil"
	}

	return fmt.Sprintf("%+v", rsp)
}
