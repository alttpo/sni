//go:build notray

package tray

import (
	"fmt"
	"sni/devices"
)

func Init() (err error) {
	return
}

func CreateSystray() {
	// sleep the main goroutine so the process does not exit immediately:
	select {}
}

func ShowMessage(appName, title, msg string) {
	fmt.Println(appName)
	fmt.Println(title)
	fmt.Println(msg)
}

func UpdateDeviceList(descriptors []devices.DeviceDescriptor) {
	// no-op
}
