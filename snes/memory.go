package snes

import (
	"context"
	"sni/protos/sni"
)

type MemoryReadRequest struct {
	RequestAddress      uint32
	RequestAddressSpace sni.AddressSpace
	RequestMapping      sni.MemoryMapping

	Size int
}

type MemoryReadResponse struct {
	MemoryReadRequest

	DeviceAddress      uint32
	DeviceAddressSpace sni.AddressSpace

	Data []byte
}

type MemoryWriteRequest struct {
	RequestAddress      uint32
	RequestAddressSpace sni.AddressSpace
	RequestMapping      sni.MemoryMapping

	Data []byte
}

type MemoryWriteResponse struct {
	RequestAddress      uint32
	RequestAddressSpace sni.AddressSpace
	RequestMapping      sni.MemoryMapping

	DeviceAddress      uint32
	DeviceAddressSpace sni.AddressSpace

	Size int
}

type DeviceMemoryUser func(ctx context.Context, memory DeviceMemory) error

type DeviceMemory interface {
	MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}
