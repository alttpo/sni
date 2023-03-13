package emunwa

import (
	"encoding/hex"
	"fmt"
	"github.com/alttpo/observable"
	"github.com/alttpo/snes/timing"
	"github.com/mitchellh/mapstructure"
	"log"
	"net"
	"net/url"
	"regexp"
	"sni/devices"
	"sni/devices/platforms"
	"sni/protos/sni"
	"sni/util"
	"sni/util/env"
	"strconv"
	"strings"
	"sync"
	"time"
)

const driverName = "emunwa"

var logDetector = false
var driver *Driver

const defaultAddressSpace = sni.AddressSpace_SnesABus

type Driver struct {
	container devices.DeviceContainer

	detectors []*Client
}

func NewDriver(addresses []*net.TCPAddr) *Driver {
	d := &Driver{
		detectors: make([]*Client, len(addresses)),
	}
	d.container = devices.NewDeviceDriverContainer(d.openDevice)

	for i, addr := range addresses {
		c := NewClient(addr, addr.String(), timing.Frame*4)
		c.MuteLog(!logDetector)
		d.detectors[i] = c
	}

	return d
}

func (d *Driver) DisplayOrder() int {
	return 1
}

func (d *Driver) DisplayName() string {
	return "EmuNWA"
}

func (d *Driver) DisplayDescription() string {
	return "Connect to a EmuNWA emulator"
}

func (d *Driver) Kind() string { return "emunwa" }

// TODO: sni.DeviceCapability_ExecuteASM
var driverCapabilities = []sni.DeviceCapability{
	sni.DeviceCapability_ReadMemory,
	sni.DeviceCapability_WriteMemory,
	sni.DeviceCapability_ResetSystem,
	sni.DeviceCapability_PauseUnpauseEmulation,
	sni.DeviceCapability_FetchFields,
	sni.DeviceCapability_NWACommand,
	sni.DeviceCapability_ReadMemoryDomain,
	sni.DeviceCapability_WriteMemoryDomain,
}

func (d *Driver) HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error) {
	return devices.CheckCapabilities(capabilities, driverCapabilities)
}

func (d *Driver) openDevice(uri *url.URL) (q devices.Device, err error) {
	// create a new device with its own connection:
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp", uri.Host)
	if err != nil {
		return
	}

	var c *Client
	c = NewClient(addr, addr.String(), time.Second*5)
	err = c.Connect()
	if err != nil {
		return
	}

	q = c
	return
}

func (d *Driver) Detect() (devs []devices.DeviceDescriptor, derr error) {
	devicesLock := sync.Mutex{}
	devs = make([]devices.DeviceDescriptor, 0, len(d.detectors))

	wg := sync.WaitGroup{}
	wg.Add(len(d.detectors))
	for i, de := range d.detectors {
		// run detectors in parallel:
		go func(i int, detector *Client) {
			defer util.Recover()

			var err error

			defer wg.Done()

			// reopen detector if necessary:
			if detector.IsClosed() {
				err = detector.Close()
				if err != nil {
					log.Printf("emunwa: error closing detector: %v\n", err)
				}
				// refresh detector:
				c := NewClient(detector.addr, fmt.Sprintf("emunwa[%d]", i), timing.Frame*4)
				c.MuteLog(!logDetector)
				d.detectors[i] = c
				detector = c
			}

			// reconnect detector if necessary:
			if !detector.IsConnected() {
				err = detector.Connect()
				if err != nil {
					if logDetector {
						log.Printf("emunwa: detect: detector[%d]: connect: %v\n", i, err)
					}
					return
				}

				// detect accidental loopback connections:
				if detector.DetectLoopback(d.detectors) {
					detector.Close()
					if logDetector {
						log.Printf("emunwa: detect: detector[%d]: loopback connection detected; breaking\n", i)
					}
					return
				}
			}

			var (
				name    string
				version string
			)

			{
				// TODO: backwards compat to EMU_INFO
				// check emulator info:
				var status []map[string]string
				var bin []byte
				bin, status, err = detector.SendCommandWaitReply("EMULATOR_INFO", time.Now().Add(timing.Frame*2))
				if err != nil {
					return
				}
				if logDetector {
					log.Printf("emunwa: detect: detector[%d]: EMULATOR_INFO\n%+v\n", i, status)
				}
				if len(status) == 0 {
					if logDetector {
						log.Printf("emunwa: detect: detector[%d]: EMULATOR_INFO did not reply properly with ASCII; instead got binary:\n%s", i, hex.Dump(bin))
					}
					return
				}
				name = status[0]["name"]
				version = status[0]["version"]
			}

			descriptor := devices.DeviceDescriptor{
				Uri:                 url.URL{Scheme: driverName, Host: detector.addr.String()},
				DisplayName:         fmt.Sprintf("%s %s (emunwa)", name, version),
				Kind:                d.Kind(),
				Capabilities:        driverCapabilities[:],
				DefaultAddressSpace: defaultAddressSpace,
				System:              "snes",
			}

			devicesLock.Lock()
			devs = append(devs, descriptor)
			devicesLock.Unlock()
		}(i, de)
	}
	wg.Wait()

	derr = nil
	return
}

func (d *Driver) DeviceKey(uri *url.URL) string {
	return uri.Host
}

func (d *Driver) Device(uri *url.URL) devices.AutoCloseableDevice {
	return devices.NewAutoCloseableDevice(d.container, uri, d.DeviceKey(uri))
}

func (d *Driver) DisconnectAll() {
	for _, deviceKey := range d.container.AllDeviceKeys() {
		device, ok := d.container.GetDevice(deviceKey)
		if ok {
			log.Printf("%s: disconnecting device '%s'\n", driverName, deviceKey)
			_ = device.Close()
			d.container.DeleteDevice(deviceKey)
		}
	}
}

type SNIMemoryDomainName = string

type coreMatchers struct {
	CoreNameRegex     *regexp.Regexp
	CoreVersionRegex  *regexp.Regexp
	CorePlatformRegex *regexp.Regexp
}
type coreDefine struct {
	Platform         string
	SNIToCoreMapping map[SNIMemoryDomainName]string
	CoreToSNIMapping map[string]SNIMemoryDomainName
}
type CoreConfig struct {
	Name    string
	Matches *coreMatchers
	Define  *coreDefine
}

var (
	coresConfig       []*CoreConfig
	currentPlatConfig *platforms.Config
)

func DriverConfig(platConfig *platforms.Config) {
	ourConfMap, ok := platConfig.Drivers["emunwa"]
	if !ok {
		log.Printf("emunwa: config: missing emunwa driver config\n")
		return
	}

	// decode emunwa driver config DTO:
	var ourConf struct {
		Cores []*struct {
			Name    string
			Matches *struct {
				CoreName     string `mapstructure:"core-name"`
				CoreVersion  string `mapstructure:"core-version"`
				CorePlatform string `mapstructure:"core-platform"`
			}
			Define *struct {
				Platform  string
				SNIToCore map[string]string `mapstructure:"sni-to-core"`
				CoreToSNI map[string]string `mapstructure:"core-to-sni"`
			}
		}
	}

	err := mapstructure.Decode(ourConfMap, &ourConf)
	if err != nil {
		log.Printf("emunwa: config: error decoding: %v\n", err)
		return
	}

	// translate config DTO to in-memory representation:
	newCoresConfig := make([]*CoreConfig, 0, len(ourConf.Cores))
	for _, coreConf := range ourConf.Cores {
		if coreConf.Matches.CoreName == "" {
			log.Printf("emunwa: config: core '%s' is missing required coreName regexp\n", coreConf.Name)
			return
		}

		// compile regexps if not empty:
		var coreNameRegex *regexp.Regexp = nil
		var coreVersionRegex *regexp.Regexp = nil
		var corePlatformRegex *regexp.Regexp = nil
		for _, m := range []struct {
			r string
			x **regexp.Regexp
			n string
		}{
			{coreConf.Matches.CoreName, &coreNameRegex, "coreName"},
			{coreConf.Matches.CoreVersion, &coreVersionRegex, "coreVersion"},
			{coreConf.Matches.CorePlatform, &corePlatformRegex, "corePlatform"},
		} {
			if m.r == "" {
				continue
			}
			*m.x, err = regexp.Compile(m.r)
			if err != nil {
				log.Printf("emunwa: config: core '%s' error compiling %s regexp `%v`: %v\n", coreConf.Name, m.n, m.r, err)
				return
			}
		}

		if _, ok = platConfig.PlatformsByName[coreConf.Define.Platform]; !ok {
			log.Printf("emunwa: config: core '%s' platform '%s' is not defined in platforms", coreConf.Name, coreConf.Define.Platform)
			return
		}

		// append the core configuration:
		newCoresConfig = append(newCoresConfig, &CoreConfig{
			Name: coreConf.Name,
			Matches: &coreMatchers{
				CoreNameRegex:     coreNameRegex,
				CoreVersionRegex:  coreVersionRegex,
				CorePlatformRegex: corePlatformRegex,
			},
			Define: &coreDefine{
				Platform:         coreConf.Define.Platform,
				SNIToCoreMapping: coreConf.Define.SNIToCore,
				CoreToSNIMapping: coreConf.Define.CoreToSNI,
			},
		})
	}

	// assign new driver config:
	coresConfig = newCoresConfig
	currentPlatConfig = platConfig
}

func DriverInit() {
	if util.IsTruthy(env.GetOrDefault("SNI_EMUNW_DISABLE", "0")) {
		log.Printf("emunwa: disabling emunwa snes driver\n")
		return
	}

	basePortStr := env.GetOrDefault("NWA_PORT_RANGE", "0xbeef")
	var basePort uint64
	var err error
	if basePort, err = strconv.ParseUint(basePortStr, 0, 16); err != nil {
		basePort = 0xbeef
		log.Printf("emunwa: unable to parse '%s', using default of 0xbeef (%d)\n", basePortStr, basePort)
	}

	disableOldRange := util.IsTruthy(env.GetOrDefault("NWA_DISABLE_OLD_RANGE", "0"))

	// comma-delimited list of host:port pairs:
	hostsStr := env.GetOrSupply("SNI_EMUNW_HOSTS", func() string {
		const count = 10
		hosts := make([]string, 0, 20)
		if disableOldRange {
			log.Printf("emunwa: disabling old port range 65400..65409 due to NWA_DISABLE_OLD_RANGE")
		}
		if disableOldRange || (basePort != 65400) {
			for i := uint64(0); i < count; i++ {
				hosts = append(hosts, fmt.Sprintf("localhost:%d", basePort+i))
			}
		}
		if !disableOldRange {
			for i := 0; i < count; i++ {
				hosts = append(hosts, fmt.Sprintf("localhost:%d", 65400+i))
			}
		}
		return strings.Join(hosts, ",")
	})

	// split the hostsStr list by commas:
	hosts := strings.Split(hostsStr, ",")

	// resolve the addresses:
	addresses := make([]*net.TCPAddr, 0, len(hosts))
	for _, host := range hosts {
		addr, err := net.ResolveTCPAddr("tcp", host)
		if err != nil {
			log.Printf("emunwa: resolve('%s'): %v\n", host, err)
			// drop the address if it doesn't resolve:
			// TODO: consider retrying the resolve later? maybe not worth worrying about.
			continue
		}

		addresses = append(addresses, addr)
	}

	if util.IsTruthy(env.GetOrDefault("SNI_EMUNW_DETECT_LOG", "0")) {
		logDetector = true
		log.Printf("emunwa: enabling emunwa detector logging")
	}

	// register the driver:
	driver = NewDriver(addresses)
	devices.Register(driverName, driver)

	// subscribe to platforms.yaml config changes:
	platforms.CurrentObs.Subscribe(observable.NewObserver("emunwa", func(event observable.Event) {
		DriverConfig(event.Value.(*platforms.Config))
	}))
}
