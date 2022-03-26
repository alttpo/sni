//go:build notray
// +build notray

package tray

import "sni/devices"

func Init() (err error) {
	return
}

func CreateSystray() {
	// sleep the main goroutine so the process does not exit immediately:
	select {}
}

func UpdateDeviceList(descriptors []devices.DeviceDescriptor) {
	// no-op
}
