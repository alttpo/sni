package emunwa

import (
	"fmt"
	"github.com/alttpo/snes/timing"
	"log"
	"net"
	"net/url"
	"sni/devices"
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
				_, status, err = detector.SendCommandWaitReply("EMULATOR_INFO", time.Now().Add(timing.Frame*2))
				if err != nil {
					return
				}
				if logDetector {
					log.Printf("emunwa: detect: detector[%d]: EMULATOR_INFO\n%+v\n", i, status)
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
}
