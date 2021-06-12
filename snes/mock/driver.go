package mock

import (
	"log"
	"sni/snes"
	"sni/util"
	"sni/util/env"
)

const driverName = "mock"

type Driver struct{}

func (d *Driver) DisplayOrder() int {
	return 1000
}

func (d *Driver) DisplayName() string {
	return "Mock Device"
}

func (d *Driver) DisplayDescription() string {
	return "Connect to a mock SNES device for testing"
}

func (d *Driver) OpenQueue(desc snes.DeviceDescriptor) (snes.Queue, error) {
	c := &Queue{}
	c.BaseInit(driverName, c)
	c.Init()
	return c, nil
}

func (d *Driver) Detect() ([]snes.DeviceDescriptor, error) {
	return []snes.DeviceDescriptor{
		&DeviceDescriptor{},
	}, nil
}

func (d *Driver) Empty() snes.DeviceDescriptor {
	return &DeviceDescriptor{}
}

func init() {
	if util.IsTruthy(env.GetOrDefault("SNI_MOCK_ENABLE", "0")) {
		log.Printf("enabling mock snes driver\n")
		snes.Register(driverName, &Driver{})
	}
}
