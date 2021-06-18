package snes

import (
	"context"
	"sni/protos/sni"
)

type MemoryReadRequest struct {
	RequestAddress      uint32
	RequestAddressSpace sni.AddressSpace

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

	Data []byte
}

type MemoryWriteResponse struct {
	RequestAddress      uint32
	RequestAddressSpace sni.AddressSpace

	DeviceAddress      uint32
	DeviceAddressSpace sni.AddressSpace

	Size int
}

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	IsClosed() bool

	// Use provides non-exclusive access to the device's subsystems to the user func
	Use(ctx context.Context, user DeviceUser) error

	// UseMemory provides exclusive access to only the memory subsystem of the device to the user func
	UseMemory(ctx context.Context, user DeviceMemoryUser) error
}

type DeviceMemoryMapping interface {
	MappingDetect(ctx context.Context, fallbackMapping *sni.MemoryMapping, inHeaderBytes []byte) (sni.MemoryMapping, bool, []byte, error)
	MappingSet(mapping sni.MemoryMapping) sni.MemoryMapping
	MappingGet() sni.MemoryMapping
}

type DeviceMemory interface {
	DeviceMemoryMapping
	MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}
