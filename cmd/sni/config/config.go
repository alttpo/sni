package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alttpo/observable"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	ConfigObservable *observable.Object
	configObservable = observable.NewObject()
	ConfigPath       string

	AppsObservable *observable.Object
	appsObservable = observable.NewObject()
	AppsPath       string
)

// configuration state:

var (
	VerboseLogging bool = false
	LogResponses   bool = false
	ShowConsole    bool = false
)

var (
	Dir    string
	Config *viper.Viper = viper.New()
	Apps   *viper.Viper = viper.New()
)

var (
	NwaDefaultPort uint64 = 0xbeef
	sniConfigs            = map[string]any{
		"debug": false,

		"grpc_listen_host":    "0.0.0.0",
		"grpc_listen_port":    8191,
		"grpcweb_listen_port": 8190,

		"usb2snes_disable":      false,
		"usb2snes_listen_addrs": "0.0.0.0:23074",
		"fxpakpro_disable":      false,

		// How long the fxpakpro driver waits on a device that has gone quiet or
		// stopped accepting data. The firmware does its FAT work inside the USB
		// interrupt handler, so a cluster allocation on a large, full, fragmented
		// card can stall for seconds before it can answer; raise these if you see
		// spurious timeouts on such a card.
		"fxpakpro_read_timeout":  "15s",
		"fxpakpro_write_timeout": "15s",
		// Whether a caller's context deadline or cancellation aborts I/O already
		// in flight to the device. With this off, a request that has already
		// started talking to the fxpakpro runs to completion (bounded by the two
		// timeouts above) rather than being cut short mid-transfer by an
		// impatient client.
		"fxpakpro_honor_caller_deadline": true,
		// Optional pause after each 512-byte chunk written during a transfer.
		// Zero, the default, writes as fast as the device accepts data. This
		// exists to test whether host write pacing affects reliability: setting
		// SNI_DEBUG=1 slows transfers dramatically (a hex dump per chunk to a
		// file and the console) and has been reported to make failing transfers
		// succeed, so this isolates the timing from the logging.
		"fxpakpro_chunk_delay": "0s",

		"retroarch_disable":    false,
		"retroarch_hosts":      "localhost:55355",
		"retroarch_detect_log": false,

		"luabridge_listen_host": "127.0.0.1",
		"luabridge_listen_port": 65398,

		"mock_enable": false,

		// sni_emunw_hosts is set dynamically when initializing the driver and initialization is conditioned on nwa_disable_old_range
		// We are not setting it here
		"emunw_disable":    false,
		"emunw_detect_log": false,

		"proxy_disable":      false,
		"proxy_backend_host": "",
	}
	nwaConfigs = map[string]any{
		"nwa_port_range":        NwaDefaultPort,
		"nwa_disable_old_range": true,
	}
	loggingConfigs = map[string]bool{
		"verboseLogging": false,
		"logResponses":   false,
	}
)

func InitDir() {
	// decide on a config directory:
	if runtime.GOOS == "windows" {
		Dir = filepath.Join(os.Getenv("LOCALAPPDATA"), "sni")
	} else {
		var err error
		Dir, err = os.UserHomeDir()
		if err != nil {
			log.Printf("could not retrieve home directory: %s\n", err)
			return
		}
		Dir = filepath.Join(Dir, ".sni")

		// Follow XDG Base Directory Specification
		if _, err := os.Stat(Dir); err != nil {
			var xdgConfig = os.Getenv("XDG_CONFIG_HOME")
			if xdgConfig == "" {
				homeDir, _ := os.UserHomeDir()
				xdgConfig = filepath.Join(homeDir, ".config")
			}
			Dir = filepath.Join(xdgConfig, "sni")
		}
	}
	// make the directory if it doesn't exist:
	_ = os.MkdirAll(Dir, 0755|os.ModeDir)
}

func Load() {
	log.Printf("config: load\n")

	loadConfig()
	loadApps()
}

func Reload() {
	ReloadConfig()
	ReloadApps()
}

func Save() {
	var err error

	log.Printf("config: save\n")

	err = Config.WriteConfigAs(ConfigPath)
	if err != nil {
		log.Printf("config: save: %s\n", err)
		return
	}
}

func loadConfig() {
	ConfigObservable = configObservable

	// load configuration:
	Config.SetEnvPrefix("SNI")
	configFilename := "config"
	Config.SetConfigName(configFilename)
	Config.SetConfigType("yaml")

	// set the path:
	ConfigPath = Dir
	Config.AddConfigPath(ConfigPath)
	ConfigPath = filepath.Join(ConfigPath, fmt.Sprintf("%s.yaml", configFilename))

	setConfigDefaults()
	// notify observers of configuration file change:
	Config.OnConfigChange(func(_ fsnotify.Event) {
		log.Printf("config: %s.yaml modified\n", configFilename)
		configObservable.Set(Config)
	})
	Config.WatchConfig()

	// reads the config file
	ReloadConfig()

	// bind environment vars so they supersede the config file
	bindConfigEnv()
}

func ReloadConfig() {
	// load configuration for the first time:
	err := Config.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no problem.
		} else {
			log.Printf("%s\n", err)
		}
		return
	}

	// publish the configuration to subscribers:
	configObservable.Set(Config)
}

func setConfigDefaults() {
	for key, value := range sniConfigs {
		Config.SetDefault(key, value)
	}

	for key, value := range nwaConfigs {
		Config.SetDefault(key, value)
	}

	for key, value := range loggingConfigs {
		Config.SetDefault(key, value)
	}
}

func bindConfigEnv() {
	// load Env Variables
	// configs with env starting with SNI_
	for key := range sniConfigs {
		err := Config.BindEnv(key)
		if err != nil {
			log.Printf("Error Binding environment variable %v: %v\n", key, err)
		}
	}

	// As stated previously, the variable associated with SNI_EMUNW_HOSTS it set dynamically later, if not bound bound in this stage
	err := Config.BindEnv("emunw_hosts")
	if err != nil {
		log.Printf("Error Binding environment variable SNI_EMUNW_HOSTS: %v\n", err)
	}

	/*
	* Parse NWA related env variable, stated as not starting with "SNI_"
	* Viper BindEnv() will allow to use these even if they are set up with "SNI_"
	* In this case, for example, viper will associate both "SNI_NWA_PORT_RANGE" and "NWA_PORT_RANGE", the later taking precedance
	 */
	for key := range nwaConfigs {
		err := Config.BindEnv(key, strings.ToUpper(key))
		if err != nil {
			log.Printf("Error Binding environment variable %v: %v\n", key, err)
		}
	}
}

func loadApps() {
	AppsObservable = appsObservable

	// load configuration:
	appsFilename := "apps"
	Apps.SetConfigName(appsFilename)
	Apps.SetConfigType("yaml")

	// set the path:
	AppsPath = Dir
	Apps.AddConfigPath(AppsPath)
	AppsPath = filepath.Join(AppsPath, fmt.Sprintf("%s.yaml", appsFilename))

	Apps.OnConfigChange(func(_ fsnotify.Event) {
		log.Printf("config: %s.yaml modified\n", appsFilename)
		ReloadApps()
	})
	Apps.WatchConfig()

	ReloadApps()
}

func ReloadApps() {
	err := Apps.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no problem.
		} else {
			log.Printf("%s\n", err)
		}
		return
	}

	appsObservable.Set(Apps)
}
