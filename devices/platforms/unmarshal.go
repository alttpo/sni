package platforms

import (
	"bytes"
	_ "embed"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"strings"
)

//go:embed platforms.yaml
var DefaultPlatformsYaml []byte

func Unmarshal(confMap map[string]interface{}) (config *Config, err error) {
	err = mapstructure.Decode(confMap, &config)
	if err != nil {
		config = nil
		return
	}

	// build platform lookup by name:
	config.ByName = make(map[string]*PlatformConf)
	for _, p := range config.Platforms {
		platformNameLower := strings.ToLower(p.Name)
		config.ByName[platformNameLower] = p

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

func LoadDefaults() (config *Config, err error) {
	v := viper.New()

	err = v.ReadConfig(bytes.NewReader(DefaultPlatformsYaml))
	if err != nil {
		return
	}

	config, err = Unmarshal(v.AllSettings())
	if err != nil {
		config = nil
		return
	}

	return
}