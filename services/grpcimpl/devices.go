package grpcimpl

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/status"
	"sni/cmd/sni/tray"
	"sni/devices"
	"sni/protos/sni"
)

func grpcError(err error) error {
	var coded *devices.CodedError
	if errors.As(err, &coded) {
		return status.Error(coded.Code, coded.Error())
	}
	return err
}

type DevicesService struct {
	sni.UnimplementedDevicesServer
}

func (s *DevicesService) ListDevices(ctx context.Context, request *sni.DevicesRequest) (*sni.DevicesResponse, error) {
	descriptors := make([]devices.DeviceDescriptor, 0, 10)
	drivers := devices.Drivers()
	for _, driver := range drivers {
		d, err := driver.Driver.Detect()
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, d...)
	}

	go tray.UpdateDeviceList(descriptors)

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

	devs := make([]*sni.DevicesResponse_Device, 0, 10)
	for _, descriptor := range descriptors {
		if !kindPredicate(descriptor.Kind) {
			continue
		}

		devs = append(devs, &sni.DevicesResponse_Device{
			Uri:                 descriptor.Uri.String(),
			DisplayName:         descriptor.DisplayName,
			Kind:                descriptor.Kind,
			Capabilities:        descriptor.Capabilities,
			DefaultAddressSpace: descriptor.DefaultAddressSpace,
		})
	}

	return &sni.DevicesResponse{Devices: devs}, nil
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
