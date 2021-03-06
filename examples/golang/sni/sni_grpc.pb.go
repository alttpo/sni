// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package sni

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DevicesClient is the client API for Devices service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DevicesClient interface {
	// detect and list devices currently connected to the system:
	ListDevices(ctx context.Context, in *DevicesRequest, opts ...grpc.CallOption) (*DevicesResponse, error)
}

type devicesClient struct {
	cc grpc.ClientConnInterface
}

func NewDevicesClient(cc grpc.ClientConnInterface) DevicesClient {
	return &devicesClient{cc}
}

func (c *devicesClient) ListDevices(ctx context.Context, in *DevicesRequest, opts ...grpc.CallOption) (*DevicesResponse, error) {
	out := new(DevicesResponse)
	err := c.cc.Invoke(ctx, "/Devices/ListDevices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DevicesServer is the server API for Devices service.
// All implementations must embed UnimplementedDevicesServer
// for forward compatibility
type DevicesServer interface {
	// detect and list devices currently connected to the system:
	ListDevices(context.Context, *DevicesRequest) (*DevicesResponse, error)
	mustEmbedUnimplementedDevicesServer()
}

// UnimplementedDevicesServer must be embedded to have forward compatible implementations.
type UnimplementedDevicesServer struct {
}

func (UnimplementedDevicesServer) ListDevices(context.Context, *DevicesRequest) (*DevicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDevices not implemented")
}
func (UnimplementedDevicesServer) mustEmbedUnimplementedDevicesServer() {}

// UnsafeDevicesServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DevicesServer will
// result in compilation errors.
type UnsafeDevicesServer interface {
	mustEmbedUnimplementedDevicesServer()
}

func RegisterDevicesServer(s grpc.ServiceRegistrar, srv DevicesServer) {
	s.RegisterService(&Devices_ServiceDesc, srv)
}

func _Devices_ListDevices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DevicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DevicesServer).ListDevices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Devices/ListDevices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DevicesServer).ListDevices(ctx, req.(*DevicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Devices_ServiceDesc is the grpc.ServiceDesc for Devices service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Devices_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Devices",
	HandlerType: (*DevicesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListDevices",
			Handler:    _Devices_ListDevices_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sni.proto",
}

// DeviceControlClient is the client API for DeviceControl service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DeviceControlClient interface {
	// only available if DeviceCapability ResetSystem is present
	ResetSystem(ctx context.Context, in *ResetSystemRequest, opts ...grpc.CallOption) (*ResetSystemResponse, error)
	// only available if DeviceCapability PauseUnpauseEmulation is present
	PauseUnpauseEmulation(ctx context.Context, in *PauseEmulationRequest, opts ...grpc.CallOption) (*PauseEmulationResponse, error)
	// only available if DeviceCapability PauseToggleEmulation is present
	PauseToggleEmulation(ctx context.Context, in *PauseToggleEmulationRequest, opts ...grpc.CallOption) (*PauseToggleEmulationResponse, error)
}

type deviceControlClient struct {
	cc grpc.ClientConnInterface
}

func NewDeviceControlClient(cc grpc.ClientConnInterface) DeviceControlClient {
	return &deviceControlClient{cc}
}

func (c *deviceControlClient) ResetSystem(ctx context.Context, in *ResetSystemRequest, opts ...grpc.CallOption) (*ResetSystemResponse, error) {
	out := new(ResetSystemResponse)
	err := c.cc.Invoke(ctx, "/DeviceControl/ResetSystem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceControlClient) PauseUnpauseEmulation(ctx context.Context, in *PauseEmulationRequest, opts ...grpc.CallOption) (*PauseEmulationResponse, error) {
	out := new(PauseEmulationResponse)
	err := c.cc.Invoke(ctx, "/DeviceControl/PauseUnpauseEmulation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceControlClient) PauseToggleEmulation(ctx context.Context, in *PauseToggleEmulationRequest, opts ...grpc.CallOption) (*PauseToggleEmulationResponse, error) {
	out := new(PauseToggleEmulationResponse)
	err := c.cc.Invoke(ctx, "/DeviceControl/PauseToggleEmulation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeviceControlServer is the server API for DeviceControl service.
// All implementations must embed UnimplementedDeviceControlServer
// for forward compatibility
type DeviceControlServer interface {
	// only available if DeviceCapability ResetSystem is present
	ResetSystem(context.Context, *ResetSystemRequest) (*ResetSystemResponse, error)
	// only available if DeviceCapability PauseUnpauseEmulation is present
	PauseUnpauseEmulation(context.Context, *PauseEmulationRequest) (*PauseEmulationResponse, error)
	// only available if DeviceCapability PauseToggleEmulation is present
	PauseToggleEmulation(context.Context, *PauseToggleEmulationRequest) (*PauseToggleEmulationResponse, error)
	mustEmbedUnimplementedDeviceControlServer()
}

// UnimplementedDeviceControlServer must be embedded to have forward compatible implementations.
type UnimplementedDeviceControlServer struct {
}

func (UnimplementedDeviceControlServer) ResetSystem(context.Context, *ResetSystemRequest) (*ResetSystemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetSystem not implemented")
}
func (UnimplementedDeviceControlServer) PauseUnpauseEmulation(context.Context, *PauseEmulationRequest) (*PauseEmulationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PauseUnpauseEmulation not implemented")
}
func (UnimplementedDeviceControlServer) PauseToggleEmulation(context.Context, *PauseToggleEmulationRequest) (*PauseToggleEmulationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PauseToggleEmulation not implemented")
}
func (UnimplementedDeviceControlServer) mustEmbedUnimplementedDeviceControlServer() {}

// UnsafeDeviceControlServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DeviceControlServer will
// result in compilation errors.
type UnsafeDeviceControlServer interface {
	mustEmbedUnimplementedDeviceControlServer()
}

func RegisterDeviceControlServer(s grpc.ServiceRegistrar, srv DeviceControlServer) {
	s.RegisterService(&DeviceControl_ServiceDesc, srv)
}

func _DeviceControl_ResetSystem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResetSystemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceControlServer).ResetSystem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceControl/ResetSystem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceControlServer).ResetSystem(ctx, req.(*ResetSystemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceControl_PauseUnpauseEmulation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PauseEmulationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceControlServer).PauseUnpauseEmulation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceControl/PauseUnpauseEmulation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceControlServer).PauseUnpauseEmulation(ctx, req.(*PauseEmulationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceControl_PauseToggleEmulation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PauseToggleEmulationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceControlServer).PauseToggleEmulation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceControl/PauseToggleEmulation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceControlServer).PauseToggleEmulation(ctx, req.(*PauseToggleEmulationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DeviceControl_ServiceDesc is the grpc.ServiceDesc for DeviceControl service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DeviceControl_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DeviceControl",
	HandlerType: (*DeviceControlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ResetSystem",
			Handler:    _DeviceControl_ResetSystem_Handler,
		},
		{
			MethodName: "PauseUnpauseEmulation",
			Handler:    _DeviceControl_PauseUnpauseEmulation_Handler,
		},
		{
			MethodName: "PauseToggleEmulation",
			Handler:    _DeviceControl_PauseToggleEmulation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sni.proto",
}

// DeviceMemoryClient is the client API for DeviceMemory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DeviceMemoryClient interface {
	// detect the current memory mapping for the given device by reading $00:FFB0 header:
	MappingDetect(ctx context.Context, in *DetectMemoryMappingRequest, opts ...grpc.CallOption) (*DetectMemoryMappingResponse, error)
	// read a single memory segment with a given size from the given device:
	SingleRead(ctx context.Context, in *SingleReadMemoryRequest, opts ...grpc.CallOption) (*SingleReadMemoryResponse, error)
	// write a single memory segment with given data to the given device:
	SingleWrite(ctx context.Context, in *SingleWriteMemoryRequest, opts ...grpc.CallOption) (*SingleWriteMemoryResponse, error)
	// read multiple memory segments with given sizes from the given device:
	MultiRead(ctx context.Context, in *MultiReadMemoryRequest, opts ...grpc.CallOption) (*MultiReadMemoryResponse, error)
	// write multiple memory segments with given data to the given device:
	MultiWrite(ctx context.Context, in *MultiWriteMemoryRequest, opts ...grpc.CallOption) (*MultiWriteMemoryResponse, error)
	// stream read multiple memory segments with given sizes from the given device:
	StreamRead(ctx context.Context, opts ...grpc.CallOption) (DeviceMemory_StreamReadClient, error)
	// stream write multiple memory segments with given data to the given device:
	StreamWrite(ctx context.Context, opts ...grpc.CallOption) (DeviceMemory_StreamWriteClient, error)
}

type deviceMemoryClient struct {
	cc grpc.ClientConnInterface
}

func NewDeviceMemoryClient(cc grpc.ClientConnInterface) DeviceMemoryClient {
	return &deviceMemoryClient{cc}
}

func (c *deviceMemoryClient) MappingDetect(ctx context.Context, in *DetectMemoryMappingRequest, opts ...grpc.CallOption) (*DetectMemoryMappingResponse, error) {
	out := new(DetectMemoryMappingResponse)
	err := c.cc.Invoke(ctx, "/DeviceMemory/MappingDetect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceMemoryClient) SingleRead(ctx context.Context, in *SingleReadMemoryRequest, opts ...grpc.CallOption) (*SingleReadMemoryResponse, error) {
	out := new(SingleReadMemoryResponse)
	err := c.cc.Invoke(ctx, "/DeviceMemory/SingleRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceMemoryClient) SingleWrite(ctx context.Context, in *SingleWriteMemoryRequest, opts ...grpc.CallOption) (*SingleWriteMemoryResponse, error) {
	out := new(SingleWriteMemoryResponse)
	err := c.cc.Invoke(ctx, "/DeviceMemory/SingleWrite", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceMemoryClient) MultiRead(ctx context.Context, in *MultiReadMemoryRequest, opts ...grpc.CallOption) (*MultiReadMemoryResponse, error) {
	out := new(MultiReadMemoryResponse)
	err := c.cc.Invoke(ctx, "/DeviceMemory/MultiRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceMemoryClient) MultiWrite(ctx context.Context, in *MultiWriteMemoryRequest, opts ...grpc.CallOption) (*MultiWriteMemoryResponse, error) {
	out := new(MultiWriteMemoryResponse)
	err := c.cc.Invoke(ctx, "/DeviceMemory/MultiWrite", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceMemoryClient) StreamRead(ctx context.Context, opts ...grpc.CallOption) (DeviceMemory_StreamReadClient, error) {
	stream, err := c.cc.NewStream(ctx, &DeviceMemory_ServiceDesc.Streams[0], "/DeviceMemory/StreamRead", opts...)
	if err != nil {
		return nil, err
	}
	x := &deviceMemoryStreamReadClient{stream}
	return x, nil
}

type DeviceMemory_StreamReadClient interface {
	Send(*MultiReadMemoryRequest) error
	Recv() (*MultiReadMemoryResponse, error)
	grpc.ClientStream
}

type deviceMemoryStreamReadClient struct {
	grpc.ClientStream
}

func (x *deviceMemoryStreamReadClient) Send(m *MultiReadMemoryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *deviceMemoryStreamReadClient) Recv() (*MultiReadMemoryResponse, error) {
	m := new(MultiReadMemoryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *deviceMemoryClient) StreamWrite(ctx context.Context, opts ...grpc.CallOption) (DeviceMemory_StreamWriteClient, error) {
	stream, err := c.cc.NewStream(ctx, &DeviceMemory_ServiceDesc.Streams[1], "/DeviceMemory/StreamWrite", opts...)
	if err != nil {
		return nil, err
	}
	x := &deviceMemoryStreamWriteClient{stream}
	return x, nil
}

type DeviceMemory_StreamWriteClient interface {
	Send(*MultiWriteMemoryRequest) error
	Recv() (*MultiWriteMemoryResponse, error)
	grpc.ClientStream
}

type deviceMemoryStreamWriteClient struct {
	grpc.ClientStream
}

func (x *deviceMemoryStreamWriteClient) Send(m *MultiWriteMemoryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *deviceMemoryStreamWriteClient) Recv() (*MultiWriteMemoryResponse, error) {
	m := new(MultiWriteMemoryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DeviceMemoryServer is the server API for DeviceMemory service.
// All implementations must embed UnimplementedDeviceMemoryServer
// for forward compatibility
type DeviceMemoryServer interface {
	// detect the current memory mapping for the given device by reading $00:FFB0 header:
	MappingDetect(context.Context, *DetectMemoryMappingRequest) (*DetectMemoryMappingResponse, error)
	// read a single memory segment with a given size from the given device:
	SingleRead(context.Context, *SingleReadMemoryRequest) (*SingleReadMemoryResponse, error)
	// write a single memory segment with given data to the given device:
	SingleWrite(context.Context, *SingleWriteMemoryRequest) (*SingleWriteMemoryResponse, error)
	// read multiple memory segments with given sizes from the given device:
	MultiRead(context.Context, *MultiReadMemoryRequest) (*MultiReadMemoryResponse, error)
	// write multiple memory segments with given data to the given device:
	MultiWrite(context.Context, *MultiWriteMemoryRequest) (*MultiWriteMemoryResponse, error)
	// stream read multiple memory segments with given sizes from the given device:
	StreamRead(DeviceMemory_StreamReadServer) error
	// stream write multiple memory segments with given data to the given device:
	StreamWrite(DeviceMemory_StreamWriteServer) error
	mustEmbedUnimplementedDeviceMemoryServer()
}

// UnimplementedDeviceMemoryServer must be embedded to have forward compatible implementations.
type UnimplementedDeviceMemoryServer struct {
}

func (UnimplementedDeviceMemoryServer) MappingDetect(context.Context, *DetectMemoryMappingRequest) (*DetectMemoryMappingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MappingDetect not implemented")
}
func (UnimplementedDeviceMemoryServer) SingleRead(context.Context, *SingleReadMemoryRequest) (*SingleReadMemoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SingleRead not implemented")
}
func (UnimplementedDeviceMemoryServer) SingleWrite(context.Context, *SingleWriteMemoryRequest) (*SingleWriteMemoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SingleWrite not implemented")
}
func (UnimplementedDeviceMemoryServer) MultiRead(context.Context, *MultiReadMemoryRequest) (*MultiReadMemoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MultiRead not implemented")
}
func (UnimplementedDeviceMemoryServer) MultiWrite(context.Context, *MultiWriteMemoryRequest) (*MultiWriteMemoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MultiWrite not implemented")
}
func (UnimplementedDeviceMemoryServer) StreamRead(DeviceMemory_StreamReadServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamRead not implemented")
}
func (UnimplementedDeviceMemoryServer) StreamWrite(DeviceMemory_StreamWriteServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamWrite not implemented")
}
func (UnimplementedDeviceMemoryServer) mustEmbedUnimplementedDeviceMemoryServer() {}

// UnsafeDeviceMemoryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DeviceMemoryServer will
// result in compilation errors.
type UnsafeDeviceMemoryServer interface {
	mustEmbedUnimplementedDeviceMemoryServer()
}

func RegisterDeviceMemoryServer(s grpc.ServiceRegistrar, srv DeviceMemoryServer) {
	s.RegisterService(&DeviceMemory_ServiceDesc, srv)
}

func _DeviceMemory_MappingDetect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DetectMemoryMappingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceMemoryServer).MappingDetect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceMemory/MappingDetect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceMemoryServer).MappingDetect(ctx, req.(*DetectMemoryMappingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceMemory_SingleRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SingleReadMemoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceMemoryServer).SingleRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceMemory/SingleRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceMemoryServer).SingleRead(ctx, req.(*SingleReadMemoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceMemory_SingleWrite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SingleWriteMemoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceMemoryServer).SingleWrite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceMemory/SingleWrite",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceMemoryServer).SingleWrite(ctx, req.(*SingleWriteMemoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceMemory_MultiRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MultiReadMemoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceMemoryServer).MultiRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceMemory/MultiRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceMemoryServer).MultiRead(ctx, req.(*MultiReadMemoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceMemory_MultiWrite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MultiWriteMemoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceMemoryServer).MultiWrite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DeviceMemory/MultiWrite",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceMemoryServer).MultiWrite(ctx, req.(*MultiWriteMemoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceMemory_StreamRead_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DeviceMemoryServer).StreamRead(&deviceMemoryStreamReadServer{stream})
}

type DeviceMemory_StreamReadServer interface {
	Send(*MultiReadMemoryResponse) error
	Recv() (*MultiReadMemoryRequest, error)
	grpc.ServerStream
}

type deviceMemoryStreamReadServer struct {
	grpc.ServerStream
}

func (x *deviceMemoryStreamReadServer) Send(m *MultiReadMemoryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *deviceMemoryStreamReadServer) Recv() (*MultiReadMemoryRequest, error) {
	m := new(MultiReadMemoryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _DeviceMemory_StreamWrite_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DeviceMemoryServer).StreamWrite(&deviceMemoryStreamWriteServer{stream})
}

type DeviceMemory_StreamWriteServer interface {
	Send(*MultiWriteMemoryResponse) error
	Recv() (*MultiWriteMemoryRequest, error)
	grpc.ServerStream
}

type deviceMemoryStreamWriteServer struct {
	grpc.ServerStream
}

func (x *deviceMemoryStreamWriteServer) Send(m *MultiWriteMemoryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *deviceMemoryStreamWriteServer) Recv() (*MultiWriteMemoryRequest, error) {
	m := new(MultiWriteMemoryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DeviceMemory_ServiceDesc is the grpc.ServiceDesc for DeviceMemory service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DeviceMemory_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DeviceMemory",
	HandlerType: (*DeviceMemoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MappingDetect",
			Handler:    _DeviceMemory_MappingDetect_Handler,
		},
		{
			MethodName: "SingleRead",
			Handler:    _DeviceMemory_SingleRead_Handler,
		},
		{
			MethodName: "SingleWrite",
			Handler:    _DeviceMemory_SingleWrite_Handler,
		},
		{
			MethodName: "MultiRead",
			Handler:    _DeviceMemory_MultiRead_Handler,
		},
		{
			MethodName: "MultiWrite",
			Handler:    _DeviceMemory_MultiWrite_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamRead",
			Handler:       _DeviceMemory_StreamRead_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamWrite",
			Handler:       _DeviceMemory_StreamWrite_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "sni.proto",
}
