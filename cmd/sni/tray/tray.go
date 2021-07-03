package tray

import (
	"fmt"
	"github.com/getlantern/systray"
	"log"
	"sni/cmd/sni/icon"
	"sni/snes"
)

var VerboseLogging bool = false

const maxItems = 10
var deviceMenuItems [maxItems]*systray.MenuItem

func UpdateDeviceList(descriptors []snes.DeviceDescriptor) {
	n := len(descriptors)
	if n > maxItems {
		n = maxItems
	}

	for i, desc := range descriptors[0:n] {
		deviceMenuItems[i].SetTitle(desc.DisplayName)
		deviceMenuItems[i].SetTooltip(desc.Kind)
		deviceMenuItems[i].Check()
		deviceMenuItems[i].Show()
	}
	for i := n; i < maxItems; i++ {
		deviceMenuItems[i].Hide()
	}
}

var (
	version string
	commit  string
	date    string
	builtBy string
)

func CreateSystray(versionParam, commitParam, dateParam, builtByParam string) {
	version = versionParam
	commit = commitParam
	date = dateParam
	builtBy = builtByParam

	// Start up a systray:
	systray.Run(trayStart, trayExit)
}

func quitSystray() {
	systray.Quit()
}

func trayExit() {
	fmt.Println("Finished quitting")
}

func trayStart() {
	// Set up the systray:
	systray.SetIcon(icon.Data)

	versionText := fmt.Sprintf("Super Nintendo Interface %s (%s)", version, commit)
	systray.SetTooltip(versionText)

	versionTooltip := fmt.Sprintf("SNI %s (%s) built on %s", version, commit, date)
	systray.AddMenuItem(versionText, versionTooltip)
	systray.AddSeparator()
	devicesMenu := systray.AddMenuItem("Devices", "Devices currently detected")
	systray.AddSeparator()
	disconnectAll := systray.AddMenuItem("Disconnect SNES", "Disconnect from all connected SNES devices")
	systray.AddSeparator()
	toggleVerbose := systray.AddMenuItemCheckbox("Log all requests", "Enable logging of all incoming requests", VerboseLogging)
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit")

	refresh := devicesMenu.AddSubMenuItem("Refresh", "Refresh list of devices")
	for i := range deviceMenuItems {
		deviceMenuItems[i] = devicesMenu.AddSubMenuItemCheckbox("_", "_", false)
		deviceMenuItems[i].Hide()
	}

	// Menu item click handler:
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				break
			case <-disconnectAll.ClickedCh:
				for _, named := range snes.Drivers() {
					log.Printf("%s: disconnecting all devices...\n", named.Name)
					named.Driver.DisconnectAll()
				}
				break
			case <-toggleVerbose.ClickedCh:
				VerboseLogging = !VerboseLogging
				if VerboseLogging {
					log.Println("enable verbose logging")
					toggleVerbose.Check()
				} else {
					log.Println("disable verbose logging")
					toggleVerbose.Uncheck()
				}
				break
			case <-refresh.ClickedCh:
				descriptors := make([]snes.DeviceDescriptor, 0, 10)
				for _, named := range snes.Drivers() {
					d, err := named.Driver.Detect()
					if err != nil {
						continue
					}

					descriptors = append(descriptors, d...)
				}
				UpdateDeviceList(descriptors)

				break
			}
		}
	}()
}
