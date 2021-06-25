package mock

import (
	"log"
	"net/url"
	"sni/protos/sni"
	"sni/snes"
	"sni/util"
	"sni/util/env"
)

const driverName = "mock"

type Driver struct {
	base snes.BaseDeviceDriver
}

var driver *Driver

func (d *Driver) DisplayOrder() int {
	return 1000
}

func (d *Driver) DisplayName() string {
	return "Mock Device"
}

func (d *Driver) DisplayDescription() string {
	return "Connect to a mock SNES device for testing"
}

func (d *Driver) Kind() string { return "mock" }

var driverCapabilities = []sni.DeviceCapability{
	sni.DeviceCapability_ReadMemory,
	sni.DeviceCapability_WriteMemory,
}

func (d *Driver) HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error) {
	return snes.CheckCapabilities(capabilities, driverCapabilities)
}

func (d *Driver) Detect() ([]snes.DeviceDescriptor, error) {
	return []snes.DeviceDescriptor{
		{
			Uri:                 url.URL{Scheme: driverName, Opaque: "mock"},
			DisplayName:         "Mock",
			Kind:                d.Kind(),
			Capabilities:        driverCapabilities[:],
			DefaultAddressSpace: sni.AddressSpace_SnesABus,
		},
	}, nil
}

func (d *Driver) openDevice(uri *url.URL) (snes.Device, error) {
	dev, ok := d.base.GetDevice(d.DeviceKey(uri))
	if ok {
		return dev, nil
	}

	mock := &Device{}
	mock.WRAM = mock.Memory[0xF50000:0xF70000]
	mock.Init()

	return mock, nil
}

func (d *Driver) Device(uri *url.URL) snes.AutoCloseableDevice {
	return snes.NewAutoCloseableDevice(
		&d.base,
		uri,
		d.DeviceKey(uri),
		d.openDevice,
	)
}

func (d *Driver) DeviceKey(uri *url.URL) string { return uri.Opaque }

func init() {
	if util.IsTruthy(env.GetOrDefault("SNI_MOCK_ENABLE", "0")) {
		log.Printf("enabling mock snes driver\n")
		driver = &Driver{}
		snes.Register(driverName, driver)
	}
}
