package config

import (
	"fmt"
	"github.com/alttpo/observable"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sni/devices/platforms"
	"strings"
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
	Dir       string
	Config    *viper.Viper = viper.New()
	Apps      *viper.Viper = viper.New()
	Platforms *viper.Viper = viper.New()
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
	}
	// make the directory if it doesn't exist:
	_ = os.MkdirAll(Dir, 0755|os.ModeDir)
}

func Load() {
	log.Printf("config: load\n")

	loadPlatforms()
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

	// notify observers of configuration file change:
	Config.OnConfigChange(func(_ fsnotify.Event) {
		log.Printf("config: %s.yaml modified\n", configFilename)
		configObservable.Set(Config)
	})
	Config.WatchConfig()

	ReloadConfig()
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

func loadPlatforms() {
	Platforms.SetConfigName("platforms")
	Platforms.SetConfigType("yaml")
	Platforms.AddConfigPath(".")
	err := Platforms.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no problem.
		} else {
			log.Printf("platforms: %s\n", err)
		}
	}

	// unmarshal the platforms section; the individual drivers unmarshal their own sections:
	confMap := Platforms.AllSettings()
	err = parsePlatformsConfigMap(confMap)
	if err != nil {
		log.Printf("platforms: %s\n", err)
		return
	}
}

func parsePlatformsConfigMap(confMap map[string]interface{}) (err error) {
	err = mapstructure.Decode(confMap, &platforms.Config)
	if err != nil {
		return
	}

	// build platform lookup by name:
	platforms.ByName = make(map[string]*platforms.PlatformConf)
	for _, p := range platforms.Config.Platforms {
		platformNameLower := strings.ToLower(p.Name)
		platforms.ByName[platformNameLower] = p

		platformNamePrefix := p.Name + "/"
		platformNamePrefixLower := platformNameLower + "/"

		for i := range p.Domains {
			name := p.Domains[i].Name
			nameLower := strings.ToLower(name)
			if !strings.HasPrefix(nameLower, platformNamePrefixLower) {
				log.Printf("platforms: WARN: domain name '%s' does not begin with '%s'", name, platformNamePrefix)
			}
		}
	}

	return
}
