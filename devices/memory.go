package devices

import (
	"context"
	"sni/protos/sni"
)

type MemoryReadRequest struct {
	RequestAddress AddressTuple

	Size int
}

type MemoryReadResponse struct {
	RequestAddress AddressTuple
	DeviceAddress  AddressTuple

	Data []byte
}

type MemoryWriteRequest struct {
	RequestAddress AddressTuple

	Data []byte
}

type MemoryWriteResponse struct {
	RequestAddress AddressTuple
	DeviceAddress  AddressTuple

	Size int
}

type DeviceMemory interface {
	DefaultAddressSpace(ctx context.Context) (space sni.AddressSpace, err error)
	MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}
