package tray

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/spf13/viper"
	"log"
	"runtime"
	"sni/cmd/sni/appversion"
	"sni/cmd/sni/config"
	"sni/cmd/sni/icon"
	"sni/cmd/sni/logging"
	"sni/ob"
	"sni/snes"
	"strings"
	"time"
)

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
		//deviceMenuItems[i].Check()
		deviceMenuItems[i].Show()
	}
	for i := n; i < maxItems; i++ {
		deviceMenuItems[i].Hide()
	}
}

func CreateSystray() {
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

	versionText := fmt.Sprintf("Super Nintendo Interface %s (%s)", appversion.Version, appversion.Commit)
	systray.SetTooltip(versionText)

	versionTooltip := fmt.Sprintf("SNI %s (%s) built on %s", appversion.Version, appversion.Commit, appversion.Date)
	versionMenuItem := systray.AddMenuItem(versionText, versionTooltip)
	systray.AddSeparator()
	devicesMenu := systray.AddMenuItem("Devices", "")
	appsMenu := systray.AddMenuItem("Applications", "")
	systray.AddSeparator()
	disconnectAll := systray.AddMenuItem("Disconnect SNES", "Disconnect from all connected SNES devices")
	systray.AddSeparator()
	toggleVerbose := systray.AddMenuItemCheckbox("Log all requests", "Enable logging of all incoming requests", config.VerboseLogging)
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit")

	// subscribe to configuration changes:
	config.ConfigObservable.Subscribe(ob.NewObserver("logging", func(object interface{}) {
		if object == nil {
			return
		}

		v, ok := object.(*viper.Viper)
		if !ok || v == nil {
			return
		}

		config.VerboseLogging = v.GetBool("verboseLogging")
		if config.VerboseLogging {
			toggleVerbose.Check()
		} else {
			toggleVerbose.Uncheck()
		}
	}))

	refresh := devicesMenu.AddSubMenuItem("Refresh", "Refresh list of devices")
	for i := range deviceMenuItems {
		deviceMenuItems[i] = devicesMenu.AddSubMenuItemCheckbox("_", "_", false)
		deviceMenuItems[i].Hide()
	}

	appsMenuItems := make([]*systray.MenuItem, 0, 10)
	appConfigs := make([]*appConfig, 0, 10)
	appsMenuTooltipNone := fmt.Sprintf("Update apps.yaml to define application shortcuts: %s", config.AppsPath)
	appsMenuTooltipSome := fmt.Sprintf("Application shortcuts defined by: %s", config.AppsPath)
	appsMenu.SetTooltip(appsMenuTooltipNone)

	appsReload := appsMenu.AddSubMenuItem("Reload Configuration", "Reload Configuration from apps.yaml")

	// subscribe to configuration changes:
	config.AppsObservable.Subscribe(ob.NewObserver("tray", func(object interface{}) {
		if object == nil {
			return
		}

		v, ok := object.(*viper.Viper)
		if !ok || v == nil {
			return
		}

		// build the apps menu:

		// parse new apps config:
		newApps := make([]*appConfig, 0, 10)
		err := v.UnmarshalKey("apps", &newApps)
		if err != nil {
			log.Printf("%s\n", err)
			return
		}

		// filter apps by OS:
		filteredApps := make([]*appConfig, 0, len(newApps))
		for _, app := range newApps {
			if app.Os != "" {
				if !strings.EqualFold(app.Os, runtime.GOOS) {
					continue
				}
			}

			filteredApps = append(filteredApps, app)
		}

		// replace:
		appConfigs = filteredApps
		if len(appConfigs) == 0 {
			appsMenu.SetTooltip(appsMenuTooltipNone)
		} else {
			appsMenu.SetTooltip(appsMenuTooltipSome)
		}

		for len(appsMenuItems) < len(appConfigs) {
			i := len(appsMenuItems)
			menuItem := appsMenu.AddSubMenuItem("", "")
			appsMenuItems = append(appsMenuItems, menuItem)

			// run a click handler goroutine for this menu item:
			go func() {
				defer func() {
					recover()
				}()

				for range menuItem.ClickedCh {
					// skip the action if this menu item no longer exists:
					if i >= len(appConfigs) {
						continue
					}

					app := appConfigs[i]
					launch(app)
				}
			}()
		}

		// set menu items:
		for i, app := range appConfigs {
			tooltip := app.Tooltip
			if tooltip == "" {
				tooltip = fmt.Sprintf("Click to launch %s at %s with args %s", app.Name, app.Path, app.Args)
			}
			appsMenuItems[i].SetTitle(app.Name)
			appsMenuItems[i].SetTooltip(tooltip)
			appsMenuItems[i].Show()
		}

		// hide extra menu items:
		for i := len(appConfigs); i < len(appsMenuItems); i++ {
			appsMenuItems[i].Hide()
		}
	}))

	// Menu item click handler:
	go func() {
		refreshPeriod := time.Tick(time.Second * 2)
		for {
			select {
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				break
			case <-versionMenuItem.ClickedCh:
				launch(&appConfig{
					Name:    "",
					Tooltip: "",
					Os:      "",
					Dir:     "",
					Path:    "",
					Args:    nil,
					Url:     logging.Dir,
				})
				break
			case <-appsReload.ClickedCh:
				config.ReloadApps()
				break
			case <-disconnectAll.ClickedCh:
				for _, named := range snes.Drivers() {
					log.Printf("%s: disconnecting all devices...\n", named.Name)
					named.Driver.DisconnectAll()
				}
				break
			case <-toggleVerbose.ClickedCh:
				config.VerboseLogging = !config.VerboseLogging
				if config.VerboseLogging {
					log.Println("enable verbose logging")
					toggleVerbose.Check()
				} else {
					log.Println("disable verbose logging")
					toggleVerbose.Uncheck()
				}
				// update config file:
				config.Config.Set("verboseLogging", config.VerboseLogging)
				config.Save()
				break
			case <-refresh.ClickedCh:
				RefreshDeviceList()
				break
			case <-refreshPeriod:
				RefreshDeviceList()
				break
			}
		}
	}()
}

func RefreshDeviceList() {
	descriptors := make([]snes.DeviceDescriptor, 0, 10)
	for _, named := range snes.Drivers() {
		d, err := named.Driver.Detect()
		if err != nil {
			continue
		}

		descriptors = append(descriptors, d...)
	}
	UpdateDeviceList(descriptors)
}
