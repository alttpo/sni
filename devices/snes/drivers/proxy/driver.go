package proxy

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sni/cmd/sni/config"
	"sni/devices"
	"sni/protos/sni"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const driverName = "proxy"

var driver *Driver

const defaultAddressSpace = sni.AddressSpace_SnesABus

// Driver implements the devices.Driver interface for proxying gRPC requests
// to a remote SNI server.
type Driver struct {
	container devices.DeviceContainer

	// backendAddr is the host:port of the remote SNI gRPC server to proxy to
	backendAddr string
}

func (d *Driver) DisplayOrder() int {
	return 100
}

func (d *Driver) DisplayName() string {
	return "SNI Proxy"
}

func (d *Driver) DisplayDescription() string {
	return "Proxy connections to a remote SNI gRPC server"
}

func (d *Driver) Kind() string { return driverName }

// driverCapabilities lists all capabilities that a proxied device could support.
// The actual capabilities are determined by the remote device and reported through Detect().
var driverCapabilities = []sni.DeviceCapability{
	sni.DeviceCapability_ReadMemory,
	sni.DeviceCapability_WriteMemory,
	sni.DeviceCapability_ResetSystem,
	sni.DeviceCapability_ResetToMenu,
	sni.DeviceCapability_PauseUnpauseEmulation,
	sni.DeviceCapability_PauseToggleEmulation,
	sni.DeviceCapability_FetchFields,
	sni.DeviceCapability_ReadDirectory,
	sni.DeviceCapability_MakeDirectory,
	sni.DeviceCapability_RemoveFile,
	sni.DeviceCapability_RenameFile,
	sni.DeviceCapability_PutFile,
	sni.DeviceCapability_GetFile,
	sni.DeviceCapability_BootFile,
	sni.DeviceCapability_NWACommand,
}

func (d *Driver) HasCapabilities(capabilities ...sni.DeviceCapability) (bool, error) {
	return devices.CheckCapabilities(capabilities, driverCapabilities)
}

func (d *Driver) Detect() (devs []devices.DeviceDescriptor, err error) {
	// connect to backend SNI server to discover its devices
	conn, cerr := grpc.NewClient(
		d.backendAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if cerr != nil {
		log.Printf("proxy: failed to connect to backend %s: %v\n", d.backendAddr, cerr)
		return nil, nil
	}
	defer conn.Close()

	client := sni.NewDevicesClient(conn)

	ctx, cancel := newDetectContext()
	defer cancel()

	resp, rerr := client.ListDevices(ctx, &sni.DevicesRequest{})
	if rerr != nil {
		log.Printf("proxy: failed to list devices from backend %s: %v\n", d.backendAddr, rerr)
		return nil, nil
	}

	devs = make([]devices.DeviceDescriptor, 0, len(resp.GetDevices()))

	for _, dev := range resp.GetDevices() {
		// build a proxy URI that encodes both the backend address and the remote device URI
		// format: proxy://<backendAddr>?uri=<remoteDeviceUri>
		proxyURI := url.URL{
			Scheme:   driverName,
			Host:     d.backendAddr,
			RawQuery: "uri=" + url.QueryEscape(dev.GetUri()),
		}

		descriptor := devices.DeviceDescriptor{
			Uri:                 proxyURI,
			DisplayName:         fmt.Sprintf("[Proxy] %s", dev.GetDisplayName()),
			Kind:                driverName,
			Capabilities:        dev.GetCapabilities(),
			DefaultAddressSpace: dev.GetDefaultAddressSpace(),
			System:              "snes",
		}

		devs = append(devs, descriptor)
	}

	return
}

func (d *Driver) DeviceKey(uri *url.URL) string {
	return uri.Host + "?" + uri.RawQuery
}

func (d *Driver) Device(uri *url.URL) devices.AutoCloseableDevice {
	return devices.NewAutoCloseableDevice(d.container, uri, d.DeviceKey(uri))
}

func (d *Driver) openDevice(uri *url.URL) (devices.Device, error) {
	backendAddr := uri.Host
	remoteURI := uri.Query().Get("uri")
	if remoteURI == "" {
		return nil, fmt.Errorf("proxy: missing 'uri' query parameter in device URI")
	}

	conn, err := grpc.NewClient(
		backendAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, devices.DeviceFatal(
			fmt.Sprintf("proxy: failed to connect to backend %s", backendAddr),
			err,
		)
	}

	dev := &Device{
		conn:       conn,
		remoteURI:  remoteURI,
		devices:    sni.NewDevicesClient(conn),
		memory:     sni.NewDeviceMemoryClient(conn),
		control:    sni.NewDeviceControlClient(conn),
		filesystem: sni.NewDeviceFilesystemClient(conn),
		info:       sni.NewDeviceInfoClient(conn),
		nwa:        sni.NewDeviceNWAClient(conn),
	}

	return dev, nil
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

func newDetectContext() (ctx_out context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func DriverInit() {
	if config.Config.GetBool("proxy_disable") {
		log.Printf("disabling proxy snes driver\n")
		return
	}

	backendAddr := config.Config.GetString("proxy_backend_host")
	if backendAddr == "" {
		// proxy driver not configured, silently skip
		return
	}

	// TODO: consider supporting an array / multiple comma-separated backend addresses in the future
	backendAddr = strings.TrimSpace(backendAddr)

	log.Printf("proxy: enabling proxy driver to backend %s\n", backendAddr)

	driver = &Driver{
		backendAddr: backendAddr,
	}
	driver.container = devices.NewDeviceDriverContainer(driver.openDevice)
	devices.Register(driverName, driver)
}
