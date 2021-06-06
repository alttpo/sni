package snes

import (
	"fmt"
	"sort"
	"sync"
)

// DeviceDescriptorBase MUST be embedded in all structs implementing DeviceDescriptor
type DeviceDescriptorBase struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// DeviceDescriptor MUST embed DeviceDescriptorBase and MAY contain extra fields used to uniquely identify a device
type DeviceDescriptor interface {
	// Base is used to fetch the DeviceDescriptorBase embedded in implementing structs
	Base() *DeviceDescriptorBase

	// GetId value is copied to the Id field of DeviceDescriptorBase
	GetId() string
	// GetDisplayName value is copied to the DisplayName field of DeviceDescriptorBase
	GetDisplayName() string
}

// MarshalDeviceDescriptor MUST be called to keep DeviceDescriptor in consistent state for marshaling
func MarshalDeviceDescriptor(device DeviceDescriptor) DeviceDescriptor {
	device.Base().Id = device.GetId()
	device.Base().DisplayName = device.GetDisplayName()
	return device
}

type Driver interface {
	// Open a connection to a specific device
	Open(desc DeviceDescriptor) (Queue, error)

	// Detect any present devices
	Detect() ([]DeviceDescriptor, error)

	// Empty Returns a descriptor with all fields empty or defaulted
	Empty() DeviceDescriptor
}

type NamedDriver struct {
	Driver Driver
	Name   string
}

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

func Open(driverName string, desc DeviceDescriptor) (Queue, error) {
	driversMu.RLock()
	driveri, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("snes: unknown driver %q (forgotten import?)", driverName)
	}

	return driveri.Open(desc)
}
