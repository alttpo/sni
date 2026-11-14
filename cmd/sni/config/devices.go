package config

import (
	"strings"
	"time"

	devconfig "github.com/alttpo/snes-devices/config"
)

// devicesConfig provides SNI's configuration to the snes-devices drivers. Settings come from SNI's viper
// configuration: the config file and SNI_* environment variables (see bindConfigEnv).
type devicesConfig struct{}

func (devicesConfig) IsSet(key string) bool { return Config.IsSet(key) }

func (devicesConfig) GetBool(key string) bool {
	// runtime toggles from the tray menu:
	switch key {
	case "verbose_logging":
		return VerboseLogging
	case "log_responses":
		return LogResponses
	}
	return Config.GetBool(key)
}

func (devicesConfig) GetString(key string) string { return Config.GetString(key) }

func (devicesConfig) GetDuration(key string) time.Duration { return Config.GetDuration(key) }

// Name returns the environment variable that sets the key.
func (devicesConfig) Name(key string) string {
	// NWA settings are also read without the SNI_ prefix:
	if _, ok := nwaConfigs[key]; ok {
		return strings.ToUpper(key)
	}
	return "SNI_" + strings.ToUpper(key)
}

// installDevicesConfig makes SNI's configuration available to the device drivers.
func installDevicesConfig() {
	devconfig.Set(devicesConfig{})
}
