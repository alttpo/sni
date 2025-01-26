package devices

import (
	"fmt"
	"net/url"
	"os"
	"sni/cmd/sni/config"
	"sni/util"
	"sort"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

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
		// use DisplayOrder iff both are available:
		if di, ok := li.(DriverDescriptor); ok {
			if dj, ok := lj.(DriverDescriptor); ok {
				return di.DisplayOrder() < dj.DisplayOrder()
			}
		}
		// fall back on Kind order:
		return li.Kind() < lj.Kind()
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

func DeviceDriverByUri(uri *url.URL) (drv Driver, err error) {
	var ok bool
	var gendrv Driver
	gendrv, ok = DriverByName(uri.Scheme)
	if !ok {
		err = fmt.Errorf("driver not found by name '%s'", uri.Scheme)
		return
	}

	drv, ok = gendrv.(Driver)
	if !ok {
		err = fmt.Errorf("driver named '%s' is not a Driver", uri.Scheme)
		return
	}

	return
}

func DeviceByUri(uri *url.URL) (driver Driver, device AutoCloseableDevice, err error) {
	driver, err = DeviceDriverByUri(uri)
	if err != nil {
		return
	}

	device = driver.Device(uri)
	return
}

// IsDisable returns a bool
// if an Environment variable is set, will return it's value
// if value is set in config file, will return it's value
// else returns default
func IsDisabled(varName string, defaultValue bool) bool {
	// read the Environment Variable
	value, present := os.LookupEnv(varName)

	if present {
		return util.IsTruthy(value)
	}

	// else return what is the config file
	// defaults to false if key not found
	if config.Config.IsSet(varName) {
		return config.Config.GetBool(varName)
	}

	return defaultValue
}
