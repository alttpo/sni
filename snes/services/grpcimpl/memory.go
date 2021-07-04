package grpcimpl

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"strings"
)

type DeviceMemoryService struct {
	sni.UnimplementedDeviceMemoryServer
}

func (s *DeviceMemoryService) MappingDetect(gctx context.Context, request *sni.DetectMemoryMappingRequest) (grsp *sni.DetectMemoryMappingResponse, gerr error) {
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

func (s *DeviceMemoryService) SingleRead(
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

func (s *DeviceMemoryService) SingleWrite(
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

func (s *DeviceMemoryService) MultiRead(
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

func (s *DeviceMemoryService) MultiWrite(
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

func (s *DeviceMemoryService) StreamRead(stream sni.DeviceMemory_StreamReadServer) error {
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

func (s *DeviceMemoryService) StreamWrite(stream sni.DeviceMemory_StreamWriteServer) error {
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

func ReadMemoryRequestString(m *sni.ReadMemoryRequest) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetRequestAddress(),
			AddressSpace:  m.GetRequestAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		m.GetSize(),
	)
}

func WriteMemoryRequestString(m *sni.WriteMemoryRequest) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetRequestAddress(),
			AddressSpace:  m.GetRequestAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		len(m.GetData()),
	)
}

func ReadMemoryResponseString(m *sni.ReadMemoryResponse) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetDeviceAddress(),
			AddressSpace:  m.GetDeviceAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		len(m.GetData()),
	)
}

func WriteMemoryResponseString(m *sni.WriteMemoryResponse) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetDeviceAddress(),
			AddressSpace:  m.GetDeviceAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		m.GetSize(),
	)
}

func (s *DeviceMemoryService) MethodRequestString(method string, req interface{}) string {
	if req == nil {
		return "nil"
	}

	switch method {
	case "/DeviceMemory/SingleRead":
		srReq := req.(*sni.SingleReadMemoryRequest)
		return fmt.Sprintf("uri:\"%s\",request:%s", srReq.GetUri(), ReadMemoryRequestString(srReq.GetRequest()))
	case "/DeviceMemory/SingleWrite":
		swReq := req.(*sni.SingleWriteMemoryRequest)
		return fmt.Sprintf("uri:\"%s\",request:%s", swReq.GetUri(), WriteMemoryRequestString(swReq.GetRequest()))
	case "/DeviceMemory/MultiRead":
		mrReq := req.(*sni.MultiReadMemoryRequest)

		sb := strings.Builder{}
		for i, rReq := range mrReq.GetRequests() {
			sb.WriteString(ReadMemoryRequestString(rReq))
			if i != len(mrReq.GetRequests())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",requests:[%s]", mrReq.GetUri(), sb.String())
	case "/DeviceMemory/MultiWrite":
		mwReq := req.(*sni.MultiWriteMemoryRequest)

		sb := strings.Builder{}
		for i, wReq := range mwReq.GetRequests() {
			sb.WriteString(WriteMemoryRequestString(wReq))
			if i != len(mwReq.GetRequests())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",requests:[%s]", mwReq.GetUri(), sb.String())
	}

	return fmt.Sprintf("%+v", req)
}

func (s *DeviceMemoryService) MethodResponseString(method string, rsp interface{}) string {
	if rsp == nil {
		return "nil"
	}

	switch method {
	case "/DeviceMemory/SingleRead":
		srReq := rsp.(*sni.SingleReadMemoryResponse)
		return fmt.Sprintf("uri:\"%s\",response:%s", srReq.GetUri(), ReadMemoryResponseString(srReq.GetResponse()))
	case "/DeviceMemory/SingleWrite":
		swReq := rsp.(*sni.SingleWriteMemoryResponse)
		return fmt.Sprintf("uri:\"%s\",response:%s", swReq.GetUri(), WriteMemoryResponseString(swReq.GetResponse()))
	case "/DeviceMemory/MultiRead":
		mrReq := rsp.(*sni.MultiReadMemoryResponse)

		sb := strings.Builder{}
		for i, rReq := range mrReq.GetResponses() {
			sb.WriteString(ReadMemoryResponseString(rReq))
			if i != len(mrReq.GetResponses())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",responses:[%s]", mrReq.GetUri(), sb.String())
	case "/DeviceMemory/MultiWrite":
		mwReq := rsp.(*sni.MultiWriteMemoryResponse)

		sb := strings.Builder{}
		for i, wReq := range mwReq.GetResponses() {
			sb.WriteString(WriteMemoryResponseString(wReq))
			if i != len(mwReq.GetResponses())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",responses:[%s]", mwReq.GetUri(), sb.String())
	}

	return fmt.Sprintf("%+v", rsp)
}
