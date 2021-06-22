package snes

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

type DeviceMemoryUser func(ctx context.Context, memory DeviceMemory) error

type DeviceMemory interface {
	MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}

type UseMemory interface {
	// UseMemory provides exclusive access to only the memory subsystem of the device to the user func
	UseMemory(ctx context.Context, requiredCapabilities []sni.DeviceCapability, user DeviceMemoryUser) error
}
