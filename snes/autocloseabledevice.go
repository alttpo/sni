package snes

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"net/url"
)

// AutoCloseableDevice is a Device wrapper that ensures that a valid Device instance is always used for every
// interface method call. On severe errors the Device is closed and reopened when needed again.
type AutoCloseableDevice interface {
	DeviceControl
	DeviceMemory
	DeviceFilesystem
}

type autoCloseableDevice struct {
	container DeviceContainer
	uri       *url.URL
	deviceKey string
}

func NewAutoCloseableDevice(container DeviceContainer, uri *url.URL, deviceKey string) AutoCloseableDevice {
	if container == nil {
		panic(fmt.Errorf("container cannot be nil"))
	}
	if uri == nil {
		panic(fmt.Errorf("uri cannot be nil"))
	}

	return &autoCloseableDevice{
		container: container,
		uri:       uri,
		deviceKey: deviceKey,
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

	// TODO: replace with errors.Is(err, snes.MustClose) check
	if device.IsClosed() {
		b.DeleteDevice(a.deviceKey)
	}
	return
}

func (a *autoCloseableDevice) ResetSystem(ctx context.Context) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		err = device.ResetSystem(ctx)
		return err
	})
	return
}

func (a *autoCloseableDevice) PauseUnpause(ctx context.Context, pausedState bool) (ok bool, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		ok, err = device.PauseUnpause(ctx, pausedState)
		return err
	})
	return
}

func (a *autoCloseableDevice) PauseToggle(ctx context.Context) (err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		err = device.PauseToggle(ctx)
		return err
	})
	return
}

func (a *autoCloseableDevice) MultiReadMemory(ctx context.Context, reads ...MemoryReadRequest) (rsp []MemoryReadResponse, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		rsp, err = device.MultiReadMemory(ctx, reads...)
		return err
	})
	return
}

func (a *autoCloseableDevice) MultiWriteMemory(ctx context.Context, writes ...MemoryWriteRequest) (rsp []MemoryWriteResponse, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		rsp, err = device.MultiWriteMemory(ctx, writes...)
		return err
	})
	return
}

func (a *autoCloseableDevice) ReadDirectory(ctx context.Context, path string) (rsp []DirEntry, err error) {
	err = a.ensureOpened(ctx, func(ctx context.Context, device Device) (err error) {
		fs, ok := device.(DeviceFilesystem)
		if !ok {
			return WithCode(codes.Unimplemented, fmt.Errorf("DeviceFilesystem not implemented"))
		}
		rsp, err = fs.ReadDirectory(ctx, path)
		return err
	})
	return
}
