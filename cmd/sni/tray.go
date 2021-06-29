package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"log"
	"sni/cmd/sni/icon"
	"sni/snes"
)

func createSystray() {
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
	disconnectAll := systray.AddMenuItem("Disconnect SNES", "Disconnect from all connected SNES devices")
	systray.AddSeparator()
	toggleVerbose := systray.AddMenuItemCheckbox("Verbose Logging", "Enable verbose logging for usb2snes server", wsVerbose)
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit")

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
				wsVerbose = !wsVerbose
				if wsVerbose {
					toggleVerbose.Check()
				} else {
					toggleVerbose.Uncheck()
				}
				break
			}
		}
	}()
}
