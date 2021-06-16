package snes

import "context"

type MemoryReadRequest struct {
	Address uint32
	Size    int
}

type MemoryReadResponse struct {
	MemoryReadRequest
	Data []byte
}

type MemoryWriteRequest struct {
	Address uint32
	Data    []byte
}

type MemoryWriteResponse struct {
	Address uint32
	Size    int
}

type DeviceMemoryUser func(context context.Context, memory DeviceMemory) error

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	IsClosed() bool

	// Use provides non-exclusive access to the device's subsystems to the user func
	Use(ctx context.Context, user DeviceUser) error

	// UseMemory provides exclusive access to only the memory subsystem of the device to the user func
	UseMemory(ctx context.Context, user DeviceMemoryUser) error
}

type DeviceMemory interface {
	MultiReadMemory(context context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(context context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}
