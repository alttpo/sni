package platforms

import (
	"github.com/alttpo/observable"
)

type PlatformConf struct {
	// every domain in the platform must start with "{platform}/"
	Name        string
	Description string
	Domains     []*DomainConf

	// computed:
	DomainsByName map[string]*DomainConf `mapstructure:"-"`
}

type DomainConf struct {
	// domain names are globally unique across all platforms
	Name string
	Size uint64
}

type Config struct {
	Platforms []*PlatformConf
	Drivers   map[string]interface{}

	// computed:
	PlatformsByName map[string]*PlatformConf `mapstructure:"-"`
}

var CurrentObs = observable.NewObject() // *platforms.Config

// Domain is a base driver-specific configuration
type Domain struct {
	DomainConf

	IsExposed      bool
	IsCoreSpecific bool

	IsReadable  bool
	IsWriteable bool
}
