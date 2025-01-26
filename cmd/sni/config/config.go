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
	// SniDebug       bool = false
	// SniDebugParsed bool = false // ShowConsole    bool = false
)

var (
	Dir    string
	Config *viper.Viper = viper.New()
	Apps   *viper.Viper = viper.New()
)

var NwaDefaultPort uint64 = 0xbeef

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

	ReloadConfig()

	// bind environment vars so they supersede the config file
	bindConfigEnv()
	// VerboseLogging = Config.GetBool("verboseLogging")
	// LogResponses = Config.GetBool("logResponses")
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
	configs := map[string]any{
		"debug": false,

		"grpc_listen_host":    "0.0.0.0",
		"grpc_listen_port":    8191,
		"grpcweb_listen_port": 8190,

		"usb2snes_disable":      false,
		"usb2snes_listen_addrs": "0.0.0.0:23074",
		"fxpakpro_disable":      false,

		"retroarch_disable":    false,
		"retroarch_hosts":      "localhost:55355",
		"retroarch_detect_log": false,

		"luabrigde_listen_host": "127.0.0.1",
		"luabrigde_listen_port": 65398,

		"emunw_disable":         false,
		"emunw_detect_log":      false,
		"nwa_port_range":        NwaDefaultPort,
		"nwa_disable_old_range": true,
	}
	for key, value := range configs {
		Config.SetDefault(key, value)
	}
}

func bindConfigEnv() {
	// load Env Variables
	// configs with env starting with sni
	sniConfigs := []string{
		"grpc_listen_host",
		"grpc_listen_port",
		"grpcweb_listen_port",
		"usb2snes_disable",
		"usb2snes_listen_addrs",
		"fxpakpro_disable",
		"debug",
		"retroarch_disable",
		"retroarch_hosts",
		"retroarch_detect_log",
		"luabrigde_listen_host",
		"luabrigde_listen_port", "emunw_disable",
		"emunw_detect_log",
		"emunw_hosts",
	}

	for _, config := range sniConfigs {
		err := Config.BindEnv(config)
		if err != nil {
			fmt.Printf("Error Binding environment variable %v: %v\n", config, err)
		}
	}

	nwaConfigs := [2]string{
		"nwa_port_range",
		"nwa_disable_old_range",
	}

	for _, config := range nwaConfigs {
		// in this case, viper will associate both "SNI_NWA_PORT_RANGE" and "NWA_PORT_RANGE", the later taking precedance
		err := Config.BindEnv(config, strings.ToUpper(config))
		if err != nil {
			fmt.Printf("Error Binding environment variable %v: %v\n", config, err)
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
