package grpcimpl

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"sni/devices"
	"sni/protos/sni"
)

type DeviceNWAService struct {
	sni.UnimplementedDeviceNWAServer
}

func (d *DeviceNWAService) NWACommand(gctx context.Context, request *sni.NWACommandRequest) (grsp *sni.NWACommandResponse, gerr error) {
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

	if _, err := driver.HasCapabilities(sni.DeviceCapability_NWACommand); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	var asciiReply []map[string]string
	var binaryReply []byte
	asciiReply, binaryReply, gerr = device.NWACommand(gctx, request.Command, request.Args, request.BinaryArg)
	if gerr != nil {
		return nil, grpcError(gerr)
	}

	// convert []map[string]string to []*NWAASCIIItem
	asciiReplyStruct := make([]*sni.NWACommandResponse_NWAASCIIItem, len(asciiReply))
	for i, m := range asciiReply {
		asciiReplyStruct[i] = &sni.NWACommandResponse_NWAASCIIItem{
			Item: m,
		}
	}

	grsp = &sni.NWACommandResponse{
		Uri:          request.Uri,
		AsciiReply:   asciiReplyStruct,
		BinaryReplay: binaryReply,
	}

	return
}
