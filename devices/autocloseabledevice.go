package devices

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"io"
	"log"
	"net/url"
	"sni/protos/sni"
	"sni/util"
	"sni/util/env"
)

// AutoCloseableDevice is a Device wrapper that ensures that a valid Device instance is always used for every
// interface method call. On severe errors the Device is closed and reopened when needed again.
type AutoCloseableDevice interface {
	io.Closer
	DeviceControl
	DeviceMemory
	DeviceFilesystem
	DeviceInfo
	DeviceNWA

	URI() *url.URL
	DeviceKey() string
}

type autoCloseableDevice struct {
	container DeviceContainer
	uri       *url.URL
	deviceKey string

	logger *log.Logger
}

var (
	sniDebug       bool
	sniDebugParsed bool
)

func NewAutoCloseableDevice(container DeviceContainer, uri *url.URL, deviceKey string) AutoCloseableDevice {
	if container == nil {
		panic(fmt.Errorf("container cannot be nil"))
	}
	if uri == nil {
		panic(fmt.Errorf("uri cannot be nil"))
	}

	if !sniDebugParsed {
		sniDebug = util.IsTruthy(env.GetOrDefault("SNI_DEBUG", "0"))
		sniDebugParsed = true
	}

	var logger *log.Logger
	if sniDebug {
		defaultLogger := log.Default()
		logger = log.New(
			defaultLogger.Writer(),
			fmt.Sprintf("autoCloseable[%s:%s]: ", uri.Scheme, deviceKey),
			defaultLogger.Flags()|log.Lmsgprefix,
		)
	}

	return &autoCloseableDevice{
		container: container,
		uri:       uri,
		deviceKey: deviceKey,
		logger:    logger,
	}
}

type deviceUser func(ctx context.Context, device Device) error

func (a *autoCloseableDevice) ensureOpened(ctx context.Context, use deviceUser) (err error) {
	b := a.container
	deviceKey := a.deviceKey

	var device Device
	device, err = b.GetOrOpenDevice(deviceKey, a.uri)
	if err != nil {
		return
	}

	err = use(ctx, device)

	// Check for fatal error and close device if so:
	if derr, ok := err.(DeviceError); ok && derr.IsFatal() {
		oerr := device.Close()
		if oerr != nil {
			log.Printf("autoCloseableDevice.ensureOpened(): device.Close(): %v\n", oerr)
		}
		b.DeleteDevice(a.deviceKey)
		return
	}

	if device.IsClosed() {
		b.DeleteDevice(a.deviceKey)
	}
	return
}

func (a *autoCloseableDevice) URI() *url.URL {
	return a.uri
}

func (a *autoCloseableDevice) DeviceKey() string {
	return a.deviceKey
}

func (a *autoCloseableDevice) Close() error {
	d, ok := a.container.GetDevice(a.deviceKey)
	if !ok {
		return nil
	}
	if a.logger != nil {
		a.logger.Printf("Close() {\n")
	}
	err := d.Close()
	if a.logger != nil {
		a.logger.Printf("Close() } -> (%#v)\n", err)
	}
	a.container.DeleteDevice(a.deviceKey)
	return err
}

func (a *autoCloseableDevice) ResetSystem(ctx context.Context) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("ResetSystem() {\n")
		}
		err = device.ResetSystem(ctx)
		if a.logger != nil {
			a.logger.Printf("ResetSystem() } -> (%#v)\n", err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) ResetToMenu(ctx context.Context) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("ResetToMenu() {\n")
		}
		err = device.ResetToMenu(ctx)
		if a.logger != nil {
			a.logger.Printf("ResetToMenu() } -> (%#v)\n", err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) PauseUnpause(ctx context.Context, pausedState bool) (ok bool, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("PauseUnpause(%#v) {\n", pausedState)
		}
		ok, err = device.PauseUnpause(ctx, pausedState)
		if a.logger != nil {
			a.logger.Printf("PauseUnpause(%#v) } -> (%#v, %#v)\n", pausedState, ok, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) PauseToggle(ctx context.Context) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("PauseToggle() {\n")
		}
		err = device.PauseToggle(ctx)
		if a.logger != nil {
			a.logger.Printf("PauseToggle() } -> (%#v)\n", err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) RequiresMemoryMappingForAddressSpace(ctx context.Context, addressSpace sni.AddressSpace) (rsp bool, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("RequiresMemoryMappingForAddressSpace(%#v) {\n", addressSpace)
		}
		rsp, err = device.RequiresMemoryMappingForAddressSpace(ctx, addressSpace)
		if a.logger != nil {
			a.logger.Printf("RequiresMemoryMappingForAddressSpace(%#v) } -> (%#v, %#v)\n", addressSpace, rsp, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) RequiresMemoryMappingForAddress(ctx context.Context, address AddressTuple) (rsp bool, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("RequiresMemoryMappingForAddress(%#v) {\n", address)
		}
		rsp, err = device.RequiresMemoryMappingForAddress(ctx, address)
		if a.logger != nil {
			a.logger.Printf("RequiresMemoryMappingForAddress(%#v) } -> (%#v, %#v)\n", address, rsp, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) (rsp []MemoryReadResponse, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("MultiReadMemory(%#v) {\n", reads)
		}
		rsp, err = device.MultiReadMemory(ctx, reads...)
		if a.logger != nil {
			a.logger.Printf("MultiReadMemory(%#v) } -> (%#v, %#v)\n", reads, rsp, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) (rsp []MemoryWriteResponse, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		if a.logger != nil {
			a.logger.Printf("MultiWriteMemory(%#v) {\n", writes)
		}
		rsp, err = device.MultiWriteMemory(ctx, writes...)
		if a.logger != nil {
			a.logger.Printf("MultiWriteMemory(%#v) } -> (%#v, %#v)\n", writes, rsp, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) FetchFields(ctx context.Context, fields ...sni.Field) (values []string, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		inf, ok := device.(DeviceInfo)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceInfo not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("FetchFields(%#v) {\n", fields)
		}
		values, err = inf.FetchFields(ctx, fields...)
		if a.logger != nil {
			a.logger.Printf("FetchFields(%#v) } -> (%#v, %#v)\n", fields, values, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) ReadDirectory(ctx context.Context, path string) (rsp []DirEntry, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("ReadDirectory(%#v) {\n", path)
		}
		rsp, err = fs.ReadDirectory(ctx, path)
		if a.logger != nil {
			a.logger.Printf("ReadDirectory(%#v) } -> (%#v, %#v)\n", path, rsp, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) MakeDirectory(ctx context.Context, path string) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("MakeDirectory(%#v) {\n", path)
		}
		err = fs.MakeDirectory(ctx, path)
		if a.logger != nil {
			a.logger.Printf("MakeDirectory(%#v) } -> (%#v)\n", path, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) RemoveFile(ctx context.Context, path string) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("RemoveFile(%#v) {\n", path)
		}
		err = fs.RemoveFile(ctx, path)
		if a.logger != nil {
			a.logger.Printf("RemoveFile(%#v) } -> (%#v)\n", path, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) RenameFile(ctx context.Context, path, newFilename string) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("RenameFile(%#v, %#v) {\n", path, newFilename)
		}
		err = fs.RenameFile(ctx, path, newFilename)
		if a.logger != nil {
			a.logger.Printf("RenameFile(%#v, %#v) } -> (%#v)\n", path, newFilename, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) PutFile(ctx context.Context, path string, size uint32, r io.Reader, progress ProgressReportFunc) (n uint32, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("PutFile(%#v, %#v) {\n", path, size)
		}
		n, err = fs.PutFile(ctx, path, size, r, progress)
		if a.logger != nil {
			a.logger.Printf("PutFile(%#v, %#v) } -> (%#v, %#v)\n", path, size, n, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) GetFile(ctx context.Context, path string, w io.Writer, sizeReceived SizeReceivedFunc, progress ProgressReportFunc) (size uint32, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("GetFile(%#v) {\n", path)
		}
		size, err = fs.GetFile(ctx, path, w, sizeReceived, progress)
		if a.logger != nil {
			a.logger.Printf("GetFile(%#v) } -> (%#v, %#v)\n", path, size, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) BootFile(ctx context.Context, path string) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("BootFile(%#v) {\n", path)
		}
		err = fs.BootFile(ctx, path)
		if a.logger != nil {
			a.logger.Printf("BootFile(%#v) } -> (%#v)\n", path, err)
		}
		return
	})
	return
}

func (a *autoCloseableDevice) NWACommand(ctx context.Context, cmd string, args string, binaryArg []byte) (asciiReply []map[string]string, binaryReply []byte, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		nwa, ok := device.(DeviceNWA)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceNWA not implemented"))
		}
		if a.logger != nil {
			a.logger.Printf("NWACommand(%#v, %#v, binary=%#v (%d bytes)) {\n", cmd, args, binaryArg != nil, len(binaryArg))
		}
		asciiReply, binaryArg, err = nwa.NWACommand(ctx, cmd, args, binaryArg)
		if a.logger != nil {
			a.logger.Printf("NWACommand(%#v, %#v, binary=%#v (%d bytes)) } -> (%#v, binary=%#v (%d bytes), %#v)\n",
				cmd, args, binaryArg != nil, len(binaryArg),
				asciiReply, binaryReply != nil, len(binaryReply),
				err)
		}
		return
	})
	return
}
