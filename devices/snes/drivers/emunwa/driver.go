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
				if detector.c != nil {
					laddr := detector.c.LocalAddr().(*net.TCPAddr)
					raddr := detector.c.RemoteAddr().(*net.TCPAddr)
					if laddr.IP.IsLoopback() && raddr.IP.IsLoopback() {
						lport := laddr.Port
						for _, ode := range d.detectors {
							if lport == ode.addr.Port {
								detector.Logf("loopback connection detected; breaking")
								detector.Close()
								return
							}
						}
					}
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
		log.Printf("disabling emunwa snes driver\n")
		return
	}

	// comma-delimited list of host:port pairs:
	hostsStr := env.GetOrSupply("SNI_EMUNW_HOSTS", func() string {
		var sb strings.Builder
		const count = 10
		for i := 0; i < count; i++ {
			sb.WriteString(fmt.Sprintf("localhost:%d", 65400+i))
			if i < count-1 {
				sb.WriteByte(',')
			}
		}
		return sb.String()
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
		log.Printf("enabling emunwa detector logging")
	}

	// register the driver:
	driver = NewDriver(addresses)
	devices.Register(driverName, driver)
}
