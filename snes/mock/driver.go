package mock

import (
	"context"
	"log"
	"net/url"
	"sni/snes"
	"sni/util"
	"sni/util/env"
	"sync"
)

const driverName = "mock"

type Driver struct {
	lock sync.Mutex
	mock *Device
}

func (d *Driver) DisplayOrder() int {
	return 1000
}

func (d *Driver) DisplayName() string {
	return "Mock Device"
}

func (d *Driver) DisplayDescription() string {
	return "Connect to a mock SNES device for testing"
}

func (d *Driver) Detect() ([]snes.DeviceDescriptor, error) {
	return []snes.DeviceDescriptor{
		{
			Uri:         url.URL{Scheme: driverName, Opaque: "mock"},
			DisplayName: "Mock",
			Kind:        "mock",
		},
	}, nil
}

func (d *Driver) OpenDevice(uri *url.URL) (snes.Device, error) {
	defer d.lock.Unlock()
	d.lock.Lock()

	return d.openDevice(uri)
}

func (d *Driver) openDevice(uri *url.URL) (snes.Device, error) {
	if d.mock == nil {
		d.mock = &Device{}
		d.mock.WRAM = d.mock.Memory[0xF50000:0xF70000]
		d.mock.Init()
	}

	return d.mock, nil
}

func (d *Driver) UseDevice(ctx context.Context, uri *url.URL, user snes.DeviceUser) (err error) {
	defer d.lock.Unlock()
	d.lock.Lock()

	var dev snes.Device
	dev, err = d.openDevice(uri)
	return user(ctx, dev)
}

func init() {
	if util.IsTruthy(env.GetOrDefault("SNI_MOCK_ENABLE", "0")) {
		log.Printf("enabling mock snes driver\n")
		snes.Register(driverName, &Driver{})
	}
}
