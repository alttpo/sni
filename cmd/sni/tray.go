package main

import (
	"fmt"
	"github.com/getlantern/systray"
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
	//systray.SetTitle("SNI")

	sniText := "SNI - Super Nintendo Interface"
	systray.SetTooltip(sniText)

	versionText := fmt.Sprintf("SNI %s", version)
	versionTooltip := fmt.Sprintf("SNI %s %s built on %s", version, commit, date)
	systray.AddMenuItem(versionText, versionTooltip)
	systray.AddMenuItem(sniText, sniText)
	systray.AddSeparator()
	disconnectAll := systray.AddMenuItem("Disconnect All Devices", "Disconnect from all SNES devices")
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
					//named.Driver.
					_ = named
				}
				break
			}
		}
	}()
}
