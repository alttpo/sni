package platforms

type PlatformConf struct {
	// every domain in the platform must start with "{platform}/"
	Name        string
	Description string
	Domains     []*DomainConf
}

type DomainConf struct {
	// domain names are globally unique across all platforms
	Name string
	Size uint64
}

type Config struct {
	Platforms []*PlatformConf
	Drivers   map[string]interface{}

	ByName map[string]*PlatformConf `mapstructure:"-"`
}

var Current *Config

// Domain is a base driver-specific configuration
type Domain struct {
	DomainConf

	IsExposed      bool
	IsCoreSpecific bool

	IsReadable  bool
	IsWriteable bool
}
