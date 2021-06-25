package snes

import (
	"context"
	"fmt"
	"net/url"
)

// AutoCloseableDevice is a Device wrapper that ensures that a valid Device instance is always used for every
// interface method call. On severe errors the Device is closed and reopened when needed again.
type AutoCloseableDevice interface {
	DeviceControl
	DeviceMemory
}

type DeviceOpener func(uri *url.URL) (Device, error)

type autoCloseableDevice struct {
	container DeviceDriverContainer
	uri       *url.URL
	deviceKey string
	opener    DeviceOpener
}

func NewAutoCloseableDevice(container DeviceDriverContainer, uri *url.URL, deviceKey string, opener DeviceOpener) AutoCloseableDevice {
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
		opener:    opener,
	}
}

type deviceUser func(ctx context.Context, device Device) error

func (a *autoCloseableDevice) ensureOpened(ctx context.Context, use deviceUser) (err error) {
	var device Device
	var ok bool

	b := a.container
	deviceKey := a.deviceKey

	device, ok = b.GetDevice(deviceKey)
	if !ok {
		device, err = b.OpenDevice(deviceKey, a.uri, a.opener)
		if err != nil {
			return
		}
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
