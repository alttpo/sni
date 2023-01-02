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

type TopLevelConf struct {
	Platforms []*PlatformConf
	Drivers   map[string]interface{}
}

var Config TopLevelConf
var ByName map[string]*PlatformConf

type Domain struct {
	DomainConf

	IsExposed      bool
	IsCoreSpecific bool

	IsReadable  bool
	IsWriteable bool
}
