package devices

import (
	"context"
	"io"
	"net/url"
	"sni/protos/sni"
)

type Driver interface {
	Kind() string

	// Detect any present devices
	Detect() ([]DeviceDescriptor, error)

	Device(uri *url.URL) AutoCloseableDevice

	DeviceKey(uri *url.URL) string

	DisconnectAll()

	HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error)
}

type DeviceDescriptor struct {
	Uri                 url.URL
	DisplayName         string
	Kind                string
	Capabilities        []sni.DeviceCapability
	DefaultAddressSpace sni.AddressSpace
	System              string
}

// Device acts as an exclusive-access gateway to the subsystems of the SNES device
type Device interface {
	io.Closer
	DeviceControl
	DeviceMemory

	IsClosed() bool
}

type DeviceMemory interface {
	RequiresMemoryMappingForAddressSpace(ctx context.Context, addressSpace sni.AddressSpace) (bool, error)
	RequiresMemoryMappingForAddress(ctx context.Context, address AddressTuple) (bool, error)
	MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) ([]MemoryReadResponse, error)
	MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) ([]MemoryWriteResponse, error)
}

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

type DeviceControl interface {
	ResetSystem(ctx context.Context) error
	ResetToMenu(ctx context.Context) error

	PauseUnpause(ctx context.Context, pausedState bool) (bool, error)
	PauseToggle(ctx context.Context) error
}

type DeviceFilesystem interface {
	ReadDirectory(ctx context.Context, path string) ([]DirEntry, error)
	MakeDirectory(ctx context.Context, path string) error
	RemoveFile(ctx context.Context, path string) error
	RenameFile(ctx context.Context, path, newFilename string) error
	PutFile(ctx context.Context, path string, size uint32, r io.Reader, progress ProgressReportFunc) (n uint32, err error)
	GetFile(ctx context.Context, path string, w io.Writer, sizeReceived SizeReceivedFunc, progress ProgressReportFunc) (size uint32, err error)
	BootFile(ctx context.Context, path string) error
}

type DirEntry struct {
	Name string
	Type sni.DirEntryType
}

type ProgressReportFunc func(current uint32, total uint32)
type SizeReceivedFunc func(size uint32)

type DeviceInfo interface {
	FetchFields(ctx context.Context, fields ...sni.Field) (values []string, err error)
}

type DeviceNWA interface {
	NWACommand(ctx context.Context, cmd string, args string, binaryArg []byte) (asciiReply []map[string]string, binaryReply []byte, err error)
}
