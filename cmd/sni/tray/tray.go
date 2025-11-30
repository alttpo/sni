//go:build !notray

package tray

import (
	"fmt"
	"log"
	"reflect"
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

type Tray struct {
	versionMenuItem *systray.MenuItem

	devicesMenu *systray.MenuItem
	appsMenu    *systray.MenuItem

	disconnectAll *systray.MenuItem

	toggleVerbose      *systray.MenuItem
	toggleLogResponses *systray.MenuItem

	toggleShowConsole *systray.MenuItem

	refresh       *systray.MenuItem
	appsMenuItems []*systray.MenuItem
	appConfigs    []*appConfig
	appsReload    *systray.MenuItem
	mQuit         *systray.MenuItem
}

var maintray Tray

func (t *Tray) updateConsole() {
	if consoleIsDynamic() {
		if t.toggleShowConsole != nil {
			if config.ShowConsole {
				t.toggleShowConsole.Check()
			} else {
				t.toggleShowConsole.Uncheck()
			}
		}
	}

	var err error
	err = consoleVisible(config.ShowConsole)
	if err != nil {
		log.Println(err)
	}
}

func (t *Tray) HandleNextAction() {
	select {
	case <-t.versionMenuItem.ClickedCh:
		launch(&appConfig{
			Name:    "",
			Tooltip: "",
			Os:      "",
			Dir:     "",
			Path:    "",
			Args:    nil,
			Url:     config.Dir,
		})
		break

	case <-t.toggleShowConsole.ClickedCh:
		config.ShowConsole = !config.ShowConsole
		// update config file:
		config.Config.Set("showConsole", config.ShowConsole)
		config.Save()
		t.updateConsole()
		break

	case <-t.refresh.ClickedCh:
		// this must not block the main thread
		RefreshDeviceList()
		break

	case <-t.appsReload.ClickedCh:
		config.ReloadApps()
		break

	case <-t.disconnectAll.ClickedCh:
		for _, named := range devices.Drivers() {
			log.Printf("%s: disconnecting all devices...\n", named.Name)
			named.Driver.DisconnectAll()
		}
		break

	case <-t.toggleVerbose.ClickedCh:
		config.VerboseLogging = !config.VerboseLogging
		if config.VerboseLogging {
			log.Println("enable verbose logging")
			t.toggleVerbose.Check()
		} else {
			log.Println("disable verbose logging")
			t.toggleVerbose.Uncheck()
		}
		// update config file:
		config.Config.Set("verboseLogging", config.VerboseLogging)
		config.Save()
		break

	case <-t.toggleLogResponses.ClickedCh:
		config.LogResponses = !config.LogResponses
		if config.LogResponses {
			log.Println("enable log responses")
			t.toggleLogResponses.Check()
		} else {
			log.Println("disable log responses")
			t.toggleLogResponses.Uncheck()
		}
		// update config file:
		config.Config.Set("logResponses", config.LogResponses)
		config.Save()
		break

	case <-t.mQuit.ClickedCh:
		log.Println("tray: requesting quit")
		systray.Quit()
		break
	}
}

func (t *Tray) Init() {
	versionText := fmt.Sprintf("Super Nintendo Interface %s (%s)", appversion.Version, appversion.Commit)
	systray.SetTooltip(versionText)

	versionTooltip := fmt.Sprintf("SNI %s (%s) built on %s", appversion.Version, appversion.Commit, appversion.Date)
	t.versionMenuItem = systray.AddMenuItem(versionText, versionTooltip)
	systray.AddSeparator()

	t.devicesMenu = systray.AddMenuItem("Devices", "")
	t.appsMenu = systray.AddMenuItem("Applications", "")
	systray.AddSeparator()

	t.disconnectAll = systray.AddMenuItem("Disconnect SNES", "Disconnect from all connected SNES devices")
	systray.AddSeparator()

	t.toggleVerbose = systray.AddMenuItemCheckbox("Log all requests", "Enable logging of all incoming requests", config.VerboseLogging)
	t.toggleLogResponses = systray.AddMenuItemCheckbox("Log all responses", "Enable logging of all outgoing response data", config.LogResponses)
	systray.AddSeparator()

	if consoleIsDynamic() {
		t.toggleShowConsole = systray.AddMenuItemCheckbox("Show Console", "Toggles visibility of console window", config.ShowConsole)
		systray.AddSeparator()
	} else {
		t.toggleShowConsole = &systray.MenuItem{
			ClickedCh: make(chan struct{}),
		}
	}
	t.mQuit = systray.AddMenuItem("Quit", "Quit")

	t.refresh = t.devicesMenu.AddSubMenuItem("Refresh", "Refresh list of devices")
	for i := range deviceMenuItems {
		deviceMenuItems[i] = t.devicesMenu.AddSubMenuItemCheckbox("_", "_", false)
		deviceMenuItems[i].Hide()
	}

	t.appsMenuItems = make([]*systray.MenuItem, 0, 10)
	t.appConfigs = make([]*appConfig, 0, 10)
	appsMenuTooltipNone := fmt.Sprintf("Update apps.yaml to define application shortcuts: %s", config.AppsPath)
	appsMenuTooltipSome := fmt.Sprintf("Application shortcuts defined by: %s", config.AppsPath)
	t.appsMenu.SetTooltip(appsMenuTooltipNone)
	t.appsReload = t.appsMenu.AddSubMenuItem("Reload Configuration", "Reload Configuration from apps.yaml")

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
		t.appConfigs = filteredApps
		if len(t.appConfigs) == 0 {
			t.appsMenu.SetTooltip(appsMenuTooltipNone)
		} else {
			t.appsMenu.SetTooltip(appsMenuTooltipSome)
		}

		for len(t.appsMenuItems) < len(t.appConfigs) {
			menuItem := t.appsMenu.AddSubMenuItem("", "")
			t.appsMenuItems = append(t.appsMenuItems, menuItem)
		}

		// set menu items:
		for i, app := range t.appConfigs {
			tooltip := app.Tooltip
			if tooltip == "" {
				tooltip = fmt.Sprintf("Click to launch %s at %s with args %s", app.Name, app.Path, app.Args)
			}
			t.appsMenuItems[i].SetTitle(app.Name)
			t.appsMenuItems[i].SetTooltip(tooltip)
			t.appsMenuItems[i].Show()
		}

		// hide extra menu items:
		for i := len(t.appConfigs); i < len(t.appsMenuItems); i++ {
			t.appsMenuItems[i].Hide()
		}
	}))
}

func (t *Tray) HandleAppMenuItems() {
	var cases []reflect.SelectCase
	for i := 0; i < len(t.appConfigs); i++ {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(t.appsMenuItems[i].ClickedCh),
			Send: reflect.Value{},
		})
	}

	i, _, ok := reflect.Select(cases)
	if !ok {
		return
	}

	// skip the action if this menu item no longer exists:
	if i >= len(t.appConfigs) {
		return
	}

	app := t.appConfigs[i]
	launch(app)
}

func trayStart() {
	// Set up the systray:
	systray.SetIcon(icon.Data)

	maintray.Init()

	// Background thread to await on systray click actions:
	go func() {
		for {
			maintray.HandleNextAction()
		}
	}()

	go func() {
		for {
			maintray.HandleAppMenuItems()
		}
	}()

	// subscribe to configuration changes:
	config.ConfigObservable.Subscribe(observable.NewObserver("logging", func(event observable.Event) {
		v, ok := event.Value.(*viper.Viper)
		if !ok || v == nil {
			return
		}

		if config.VerboseLogging {
			maintray.toggleVerbose.Check()
		} else {
			maintray.toggleVerbose.Uncheck()
		}

		if config.LogResponses {
			maintray.toggleLogResponses.Check()
		} else {
			maintray.toggleLogResponses.Uncheck()
		}

		config.ShowConsole = v.GetBool("showConsole")
		maintray.updateConsole()
	}))

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
