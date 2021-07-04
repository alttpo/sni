package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sni/ob"
)

var (
	ConfigObservable ob.Observable
	configObservable = ob.NewObservable()
	ConfigPath       string

	AppsObservable ob.Observable
	appsObservable = ob.NewObservable()
	AppsPath       string
)

var VerboseLogging bool = false

var (
	Config *viper.Viper = viper.New()
	Apps   *viper.Viper = viper.New()
)

func Load() {
	log.Printf("config: load\n")

	loadConfig()
	loadApps()
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
	if runtime.GOOS == "windows" {
		ConfigPath = os.ExpandEnv("$LOCALAPPDATA/sni/")
		_ = os.Mkdir(ConfigPath, 0644|os.ModeDir)
		Config.AddConfigPath(ConfigPath)
	} else {
		ConfigPath = os.ExpandEnv("$HOME/.sni/")
		Config.AddConfigPath(ConfigPath)
	}
	ConfigPath = filepath.Join(ConfigPath, fmt.Sprintf("%s.yaml", configFilename))

	Config.OnConfigChange(func(_ fsnotify.Event) {
		log.Printf("config: %s.yaml modified\n", configFilename)
		configObservable.ObjectPublish(Config)
	})
	Config.WatchConfig()

	err := Config.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no problem.
		} else {
			log.Printf("%s\n", err)
		}
		return
	}

	configObservable.ObjectPublish(Config)
}

func loadApps() {
	AppsObservable = appsObservable

	// load configuration:
	appsFilename := "apps"
	Apps.SetConfigName(appsFilename)
	Apps.SetConfigType("yaml")
	if runtime.GOOS == "windows" {
		AppsPath = os.ExpandEnv("$LOCALAPPDATA/sni/")
		_ = os.Mkdir(AppsPath, 0644|os.ModeDir)
		Apps.AddConfigPath(AppsPath)
	} else {
		AppsPath = os.ExpandEnv("$HOME/.sni/")
		Apps.AddConfigPath(AppsPath)
	}
	AppsPath = filepath.Join(ConfigPath, fmt.Sprintf("%s.yaml", appsFilename))

	Apps.OnConfigChange(func(_ fsnotify.Event) {
		log.Printf("config: %s.yaml modified\n", appsFilename)
		appsObservable.ObjectPublish(Apps)
	})
	Apps.WatchConfig()

	err := Apps.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no problem.
		} else {
			log.Printf("%s\n", err)
		}
		return
	}

	appsObservable.ObjectPublish(Apps)
}
