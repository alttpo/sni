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

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	IsClosed() bool

	// Use provides non-exclusive access to the device's subsystems to the user func
	Use(ctx context.Context, user DeviceUser) error

	// UseMemory provides exclusive access to only the memory subsystem of the device to the user func
	UseMemory(ctx context.Context, user DeviceMemoryUser) error

	// UseControl provides exclusive access to only the control subsystem of the device to the user func
	UseControl(ctx context.Context, user DeviceControlUser) error
}

type DeviceMemoryUser func(ctx context.Context, memory DeviceMemory) error

type DeviceMemory interface {
	MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}

type DeviceControlUser func(ctx context.Context, control DeviceControl) error

type DeviceControl interface {
	ResetSystem(ctx context.Context) error

	PauseUnpause(ctx context.Context, pausedState bool) (bool, error)
	PauseToggle(ctx context.Context) error
}
