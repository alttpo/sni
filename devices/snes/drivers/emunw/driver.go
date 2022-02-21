package emunw

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"sni/devices"
	"sni/devices/snes/timing"
	"sni/protos/sni"
	"sni/util"
	"sni/util/env"
	"strings"
	"sync"
	"time"
)

const driverName = "emunw"

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
		d.detectors[i] = c
	}

	return d
}

func (d *Driver) DisplayOrder() int {
	return 1
}

func (d *Driver) DisplayName() string {
	return "EmuNW"
}

func (d *Driver) DisplayDescription() string {
	return "Connect to a EmuNW emulator"
}

func (d *Driver) Kind() string { return "emunw" }

// TODO: sni.DeviceCapability_ExecuteASM
var driverCapabilities = []sni.DeviceCapability{
	sni.DeviceCapability_ReadMemory,
	sni.DeviceCapability_WriteMemory,
	sni.DeviceCapability_ResetSystem,
	sni.DeviceCapability_PauseUnpauseEmulation,
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

func (d *Driver) Detect() (devs []devices.DeviceDescriptor, err error) {
	devicesLock := sync.Mutex{}
	devs = make([]devices.DeviceDescriptor, 0, len(d.detectors))

	wg := sync.WaitGroup{}
	wg.Add(len(d.detectors))
	for i, de := range d.detectors {
		// run detectors in parallel:
		go func(i int, detector *Client) {
			defer wg.Done()

			// reopen detector if necessary:
			if detector.IsClosed() {
				detector.Close()
				// refresh detector:
				c := NewClient(detector.addr, fmt.Sprintf("emunw[%d]", i), timing.Frame*4)
				d.detectors[i] = c
				detector = c
			}

			// reconnect detector if necessary:
			if !detector.IsConnected() {
				err = detector.Connect()
				if err != nil {
					if logDetector {
						log.Printf("emunw: detect: detector[%d]: connect: %v\n", i, err)
					}
					return
				}
			}

			{
				// check emulator status:
				var status []map[string]string
				_, status, err = detector.SendCommandWaitReply("EMU_STATUS", time.Now().Add(timing.Frame*2))
				if err != nil {
					return
				}
				if logDetector {
					log.Printf("emunw: detect: detector[%d]:\n%+v\n", i, status)
				}
			}

			descriptor := devices.DeviceDescriptor{
				Uri:                 url.URL{Scheme: driverName, Host: detector.addr.String()},
				DisplayName:         fmt.Sprintf("EmuNW (%s)", detector.addr),
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

	err = nil
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
		log.Printf("disabling emunw snes driver\n")
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
			log.Printf("emunw: resolve('%s'): %v\n", host, err)
			// drop the address if it doesn't resolve:
			// TODO: consider retrying the resolve later? maybe not worth worrying about.
			continue
		}

		addresses = append(addresses, addr)
	}

	if util.IsTruthy(env.GetOrDefault("SNI_EMUNW_DETECT_LOG", "0")) {
		logDetector = true
		log.Printf("enabling emunw detector logging")
	}

	// register the driver:
	driver = NewDriver(addresses)
	devices.Register(driverName, driver)
}
