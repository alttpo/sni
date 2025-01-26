//go:build !notray
// +build !notray

package tray

import (
	"fmt"
	"log"
	"runtime"
	"sni/cmd/sni/appversion"
	"sni/cmd/sni/config"
	"sni/cmd/sni/icon"
	"sni/devices"
	"sni/util"
	"strings"
	"sync"
	"time"

	"github.com/alttpo/observable"
	"github.com/alttpo/systray"
	"github.com/spf13/viper"
)

const maxItems = 10

var (
	deviceMenuItemsMu sync.Mutex
	deviceMenuItems   [maxItems]*systray.MenuItem
)

func Init() (err error) {
	err = initConsole()
	return
}

func UpdateDeviceList(deviceDescriptors []devices.DeviceDescriptor) {
	defer util.Recover()

	deviceMenuItemsMu.Lock()
	n := len(deviceDescriptors)
	if n > maxItems {
		n = maxItems
	}

	for i, desc := range deviceDescriptors[0:n] {
		if deviceMenuItems[i] == nil {
			continue
		}

		deviceMenuItems[i].SetTitle(desc.DisplayName)
		deviceMenuItems[i].SetTooltip(desc.Kind)
		// deviceMenuItems[i].Check()
		deviceMenuItems[i].Show()
	}

	for i := n; i < maxItems; i++ {
		if deviceMenuItems[i] == nil {
			continue
		}

		deviceMenuItems[i].Hide()
	}
	deviceMenuItemsMu.Unlock()
}

func CreateSystray() {
	// Start up a systray:
	systray.Run(trayStart, trayExit)
	log.Println("tray: exited main loop")
}

func ShowMessage(appName, title, msg string) {
	systray.Run(
		func() {
			systray.SetIcon(icon.Data)

			versionText := fmt.Sprintf("Super Nintendo Interface %s (%s)", appversion.Version, appversion.Commit)
			systray.SetTooltip(versionText)

			// show balloon notification for 10 seconds:
			systray.ShowMessage(appName, title, msg)
			time.Sleep(10 * time.Second)

			systray.Quit()
		},
		func() {},
	)
}

func quitSystray() {
	systray.Quit()
}

func trayExit() {
	log.Println("tray: finished quitting")
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
	toggleLogResponses := systray.AddMenuItemCheckbox("Log all responses", "Enable logging of all outgoing response data", config.LogResponses)
	systray.AddSeparator()

	var toggleShowConsole *systray.MenuItem
	updateConsole := func() {
		if toggleShowConsole != nil {
			if config.ShowConsole {
				toggleShowConsole.Check()
			} else {
				toggleShowConsole.Uncheck()
			}
		}

		var err error
		err = consoleVisible(config.ShowConsole)
		if err != nil {
			log.Println(err)
		}
	}

	if consoleIsDynamic() {
		toggleShowConsole = systray.AddMenuItemCheckbox("Show Console", "Toggles visibility of console window", config.ShowConsole)
		toggleShowConsole.ClickedFunc = func(item *systray.MenuItem) {
			go func() {
				config.ShowConsole = !config.ShowConsole
				// update config file:
				config.Config.Set("showConsole", config.ShowConsole)
				config.Save()
				updateConsole()
			}()
		}
		systray.AddSeparator()
	}
	mQuit := systray.AddMenuItem("Quit", "Quit")

	// subscribe to configuration changes:
	config.ConfigObservable.Subscribe(observable.NewObserver("logging", func(event observable.Event) {
		v, ok := event.Value.(*viper.Viper)
		if !ok || v == nil {
			return
		}

		if config.VerboseLogging {
			toggleVerbose.Check()
		} else {
			toggleVerbose.Uncheck()
		}

		if config.LogResponses {
			toggleLogResponses.Check()
		} else {
			toggleLogResponses.Uncheck()
		}

		config.ShowConsole = v.GetBool("showConsole")
		updateConsole()
	}))

	refresh := devicesMenu.AddSubMenuItem("Refresh", "Refresh list of devices")
	refresh.ClickedFunc = func(item *systray.MenuItem) {
		// this must not block the main thread
		go RefreshDeviceList()
	}
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
	config.AppsObservable.Subscribe(observable.NewObserver("tray", func(event observable.Event) {
		v, ok := event.Value.(*viper.Viper)
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
			menuItem.ClickedFunc = func(item *systray.MenuItem) {
				// skip the action if this menu item no longer exists:
				if i >= len(appConfigs) {
					return
				}

				app := appConfigs[i]
				go launch(app)
			}
			appsMenuItems = append(appsMenuItems, menuItem)
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

	// click handlers:
	versionMenuItem.ClickedFunc = func(item *systray.MenuItem) {
		go launch(&appConfig{
			Name:    "",
			Tooltip: "",
			Os:      "",
			Dir:     "",
			Path:    "",
			Args:    nil,
			Url:     config.Dir,
		})
	}

	appsReload.ClickedFunc = func(item *systray.MenuItem) {
		go config.ReloadApps()
	}

	disconnectAll.ClickedFunc = func(item *systray.MenuItem) {
		go func() {
			for _, named := range devices.Drivers() {
				log.Printf("%s: disconnecting all devices...\n", named.Name)
				named.Driver.DisconnectAll()
			}
		}()
	}

	toggleVerbose.ClickedFunc = func(item *systray.MenuItem) {
		go func() {
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
		}()
	}
	toggleLogResponses.ClickedFunc = func(item *systray.MenuItem) {
		go func() {
			config.LogResponses = !config.LogResponses
			if config.LogResponses {
				log.Println("enable log responses")
				toggleLogResponses.Check()
			} else {
				log.Println("disable log responses")
				toggleLogResponses.Uncheck()
			}
			// update config file:
			config.Config.Set("logResponses", config.LogResponses)
			config.Save()
		}()
	}

	mQuit.ClickedFunc = func(item *systray.MenuItem) {
		go func() {
			log.Println("tray: requesting quit")
			systray.Quit()
		}()
	}

	// refresh device list periodically:
	go func() {
		defer util.Recover()

		refreshPeriod := time.Tick(time.Second * 2)
		for range refreshPeriod {
			RefreshDeviceList()
		}
	}()
}

func RefreshDeviceList() {
	descriptors := make([]devices.DeviceDescriptor, 0, 10)
	for _, named := range devices.Drivers() {
		d, err := named.Driver.Detect()
		if err != nil {
			continue
		}

		descriptors = append(descriptors, d...)
	}

	UpdateDeviceList(descriptors)
}
