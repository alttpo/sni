package snes

import (
	"context"
	"fmt"
	"net/url"
	"sni/protos/sni"
	"sort"
	"sync"
)

type DeviceDescriptor struct {
	Uri                 url.URL
	DisplayName         string
	Kind                string
	Capabilities        []sni.DeviceCapability
	DefaultAddressSpace sni.AddressSpace
}

type Driver interface {
	Kind() string

	// Detect any present devices
	Detect() ([]DeviceDescriptor, error)
}

type DeviceUser func(context.Context, Device) error

type UseDevice interface {
	// UseDevice grants non-exclusive access for DeviceUser to a Device uniquely identified by its uri
	UseDevice(ctx context.Context, uri *url.URL, requiredCapabilities []sni.DeviceCapability, user DeviceUser) error
}

// DeviceDriver extends Driver
type DeviceDriver interface {
	UseDevice

	DeviceKey(uri *url.URL) string

	HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error)
}

type NamedDriver struct {
	Driver Driver
	Name   string
}

// DriverDescriptor extends Driver
type DriverDescriptor interface {
	DisplayName() string

	DisplayDescription() string

	DisplayOrder() int
}

type NamedDriverDevicePair struct {
	NamedDriver NamedDriver
	Device      DeviceDescriptor
}

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

// Register makes a SNES driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("snes: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("snes: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func unregisterAllDrivers() {
	driversMu.Lock()
	defer driversMu.Unlock()
	// For tests.
	drivers = make(map[string]Driver)
}

// Drivers returns a list of the registered drivers.
func Drivers() []NamedDriver {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]NamedDriver, 0, len(drivers))
	for name, driver := range drivers {
		list = append(list, NamedDriver{driver, name})
	}
	sort.Slice(list, func(i, j int) bool {
		li := list[i].Driver
		lj := list[j].Driver
		if di, ok := li.(DriverDescriptor); ok {
			if dj, ok := lj.(DriverDescriptor); ok {
				return di.DisplayOrder() < dj.DisplayOrder()
			}
		}
		return false
	})
	return list
}

// DriverNames returns a sorted list of the names of the registered drivers.
func DriverNames() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

func DriverByName(name string) (Driver, bool) {
	d, ok := drivers[name]
	return d, ok
}

func DeviceDriverByUri(uri *url.URL) (drv DeviceDriver, err error) {
	var ok bool
	var gendrv Driver
	gendrv, ok = DriverByName(uri.Scheme)
	if !ok {
		err = fmt.Errorf("driver not found by name '%s'", uri.Scheme)
		return
	}

	drv, ok = gendrv.(DeviceDriver)
	if !ok {
		err = fmt.Errorf("driver named '%s' is not a DeviceDriver", uri.Scheme)
		return
	}

	return
}

func WithDevice(ctx context.Context, uri *url.URL, requiredCapabilities []sni.DeviceCapability, user DeviceUser) (err error) {
	var drv DeviceDriver
	drv, err = DeviceDriverByUri(uri)
	if err != nil {
		return
	}

	return drv.UseDevice(ctx, uri, requiredCapabilities, user)
}

func WithDeviceMemory(ctx context.Context, uri *url.URL, requiredCapabilities []sni.DeviceCapability, user DeviceMemoryUser) (err error) {
	var drv DeviceDriver
	drv, err = DeviceDriverByUri(uri)
	if err != nil {
		return
	}

	return drv.UseDevice(ctx, uri, requiredCapabilities, func(ctx context.Context, device Device) error {
		return device.UseMemory(ctx, nil, user)
	})
}

func WithDeviceControl(ctx context.Context, uri *url.URL, requiredCapabilities []sni.DeviceCapability, user DeviceControlUser) (err error) {
	var drv DeviceDriver
	drv, err = DeviceDriverByUri(uri)
	if err != nil {
		return
	}

	return drv.UseDevice(ctx, uri, requiredCapabilities, func(ctx context.Context, device Device) error {
		return device.UseControl(ctx, nil, user)
	})
}
