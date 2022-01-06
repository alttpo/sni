package webdav

import (
	"context"
	"github.com/patrickmn/go-cache"
	dav "golang.org/x/net/webdav"
	"io/fs"
	"log"
	"os"
	"path"
	"sni/protos/sni"
	"sni/snes"
	"strings"
	"time"
)

// AdapterFileSystem adapts the underlying simplistic snes.DeviceFilesystem interface to the more flexible
// dav.FileSystem interface. The major limitation of the snes.DeviceFilesystem is the lack of random access
// on files. The current sd2snes/fxpakpro firmware (v1.11.0) is limited to transferring files (GET/PUT)
// sequentially from start to end without interruption. Thus, the dav.File instance returned from OpenFile
// must ensure that reading or writing is done sequentially and must not allow seeking after a read or write
// has started.
type AdapterFileSystem struct {
	statsC    *cache.Cache
	childrenC *cache.Cache

	drivers map[string]*driverDevices
}

type driverDevices struct {
	snes.NamedDriver

	devices map[string]snes.AutoCloseableDevice
}

func NewAdapterFileSystem() (a *AdapterFileSystem) {
	a = &AdapterFileSystem{
		statsC:    cache.New(time.Second, time.Second*15),
		childrenC: cache.New(time.Second, time.Second*15),
		drivers:   make(map[string]*driverDevices),
	}

	all := snes.Drivers()
	for _, d := range all {
		// make sure the driver has filesystem caps:
		ok, _ := d.Driver.HasCapabilities(sni.DeviceCapability_ReadDirectory)
		if !ok {
			continue
		}

		a.drivers[d.Name] = newDriverDevices(d)
	}

	return
}

func newDriverDevices(d snes.NamedDriver) (dd *driverDevices) {
	dd = &driverDevices{
		NamedDriver: d,
		devices:     make(map[string]snes.AutoCloseableDevice),
	}
	return
}

func (dd *driverDevices) refreshDevices() (err error) {
	drv := dd.Driver

	var detected []snes.DeviceDescriptor
	detected, err = drv.Detect()
	if err != nil {
		return
	}

	mutated := false

	// make a copy of existing device map:
	existing := make(map[string]snes.AutoCloseableDevice, len(dd.devices))
	for k, v := range dd.devices {
		existing[k] = v
	}

	refreshed := make(map[string]snes.AutoCloseableDevice, len(existing))

	// keep existing or add new devices:
	for _, desc := range detected {
		key := drv.DeviceKey(&desc.Uri)
		if dev, ok := existing[key]; ok {
			// kept:
			refreshed[key] = dev
			delete(existing, key)
		} else {
			// added:
			_, dev, err = snes.DeviceByUri(&desc.Uri)
			if err != nil {
				return
			}
			refreshed[key] = dev
			mutated = true
		}
	}

	// remove/close missing devices:
	for _, dev := range existing {
		// removed:
		err = dev.Close()
		if err != nil {
			log.Printf("webdav: refreshDevices: close: %v\n", err)
		}
		mutated = true
	}

	if mutated {
		dd.devices = refreshed
	}

	return
}

func (a *AdapterFileSystem) invalidateStat(full cleanedPath) {
	// invalidate cache:
	key := strings.ToLower(string(full))
	a.statsC.Delete(key)
	parent, _ := path.Split(key)
	a.childrenC.Delete(parent)
}

func (a *AdapterFileSystem) getStat(ctx context.Context, full cleanedPath, key string) (stat *fileInfo, err error) {
	// attempt cache read:
	o, ok := a.statsC.Get(key)
	if ok {
		// cache hit:
		stat = o.(*fileInfo)
		return
	}

	// compute result:
	stat, err = a.stat(ctx, full)
	if err != nil {
		return
	}

	// cache result:
	a.statsC.Set(key, stat, cache.DefaultExpiration)
	return
}

func (a *AdapterFileSystem) getStatChildren(ctx context.Context, full cleanedPath, key string) (children []fs.FileInfo, err error) {
	// attempt cache read:
	o, ok := a.childrenC.Get(key)
	if ok {
		// cache hit:
		children = o.([]fs.FileInfo)
		return
	}

	// compute result:
	children, err = a.statChildren(ctx, full)
	if err != nil {
		return
	}

	// cache result:
	a.childrenC.Set(key, children, cache.DefaultExpiration)
	for _, child := range children {
		fullPath := path.Join(string(full), child.Name())
		fullPathKey := strings.ToLower(fullPath)
		a.statsC.Set(fullPathKey, child, cache.DefaultExpiration)
	}
	return
}

func (a *AdapterFileSystem) newWriteable(full cleanedPath, stat *fileInfo) (f *writeable, err error) {
	driverName, deviceName, remainder := a.pathParse(full)
	if driverName == "" {
		err = fs.ErrInvalid
		return
	}
	if deviceName == "" {
		err = fs.ErrInvalid
		return
	}
	if remainder == "" {
		err = fs.ErrInvalid
		return
	}

	drv := stat.driver
	device, ok := drv.devices[deviceName]
	if !ok {
		err = fs.ErrInvalid
		return
	}

	f = &writeable{
		a:         a,
		full:      full,
		driver:    stat.driver,
		device:    device,
		remainder: remainder,
		stat:      stat,
		buf:       make([]byte, 0, 1024),
	}
	return
}

func (a *AdapterFileSystem) getDevice(driverName, deviceName string) (device snes.AutoCloseableDevice, err error) {
	if driverName != "" && deviceName != "" {
		var ok bool
		var drv *driverDevices
		drv, ok = a.drivers[driverName]
		if !ok {
			err = fs.ErrInvalid
			return
		}

		device, ok = drv.devices[deviceName]
		if !ok {
			err = fs.ErrInvalid
			return
		}
	}

	return
}

func (a *AdapterFileSystem) newReadable(full cleanedPath, stat *fileInfo, children []fs.FileInfo) (f *readable, err error) {
	var drv *driverDevices
	var device snes.AutoCloseableDevice

	driverName, deviceName, remainder := a.pathParse(full)
	if driverName != "" && deviceName != "" {
		drv = stat.driver

		var ok bool
		device, ok = drv.devices[deviceName]
		if !ok {
			err = fs.ErrInvalid
			return
		}
	}

	f = &readable{
		a:         a,
		full:      full,
		driver:    stat.driver,
		device:    device,
		remainder: remainder,
		stat:      stat,
		children:  children,
	}
	return
}

func (a *AdapterFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (f dav.File, err error) {
	log.Printf("a.OpenFile(%#v, %#v, %#v) {\n", name, flag, perm)
	defer func() {
		log.Printf("a.OpenFile(%#v, %#v, %#v) } -> (%p, %#v)\n", name, flag, perm, f, err)
	}()
	if flag&os.O_RDWR != 0 || flag&os.O_WRONLY != 0 {
		// writable open:
		full, _ := a.pathClean(name)

		var stat *fileInfo
		driverName, deviceName, remainder := a.pathParse(full)

		driver, ok := a.drivers[driverName]
		if !ok {
			err = fs.ErrInvalid
			return
		}

		_, file := path.Split(remainder)

		stat = &fileInfo{
			name:      file,
			isDir:     false,
			driver:    driver,
			deviceKey: deviceName,
		}

		f, err = a.newWriteable(full, stat)
		return
	} else {
		// readable open:
		full, key := a.pathClean(name)

		var stat *fileInfo
		stat, err = a.getStat(ctx, full, key)
		if err != nil {
			return
		}

		var children []fs.FileInfo
		if stat.IsDir() {
			children, err = a.getStatChildren(ctx, full, key)
			if err != nil {
				return
			}
		}

		f, err = a.newReadable(full, stat, children)
		return
	}
}

func (a *AdapterFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	log.Printf("a.Mkdir(%#v, %#v)\n", name, perm)

	_ = perm

	full, _ := a.pathClean(name)

	driverName, deviceName, remainder := a.pathParse(full)

	var device snes.AutoCloseableDevice
	device, err = a.getDevice(driverName, deviceName)
	if err != nil {
		return
	}

	err = device.MakeDirectory(ctx, remainder)
	a.invalidateStat(full)
	return
}

func (a *AdapterFileSystem) RemoveAll(ctx context.Context, name string) (err error) {
	log.Printf("a.RemoveAll(%#v)\n", name)

	full, _ := a.pathClean(name)

	driverName, deviceName, remainder := a.pathParse(full)

	var device snes.AutoCloseableDevice
	device, err = a.getDevice(driverName, deviceName)
	if err != nil {
		return
	}

	err = device.RemoveFile(ctx, remainder)
	a.invalidateStat(full)
	return
}

func (a *AdapterFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	log.Printf("a.Rename(%#v, %#v)\n", oldName, newName)
	return nil
}

func (a *AdapterFileSystem) Stat(ctx context.Context, name string) (stat os.FileInfo, err error) {
	log.Printf("a.Stat(%#v)\n", name)

	full, key := a.pathClean(name)
	return a.getStat(ctx, full, key)
}

type cleanedPath string

func (a *AdapterFileSystem) pathClean(name string) (full cleanedPath, key string) {
	// make sure we at least start with a /:
	if name == "" {
		name = "/"
	}

	full = cleanedPath(path.Clean(name))
	key = strings.ToLower(string(full))

	return
}

func (a *AdapterFileSystem) pathParse(full cleanedPath) (driverName, deviceName, remainder string) {
	parts := strings.Split(string(full)[1:], "/")
	if len(parts) >= 1 {
		driverName = parts[0]
	}
	if len(parts) >= 2 {
		deviceName = parts[1]
	}
	if len(parts) >= 3 {
		remainder = strings.Join(parts[2:], "/")
	}

	return
}

func (a *AdapterFileSystem) statChildren(ctx context.Context, full cleanedPath) (children []fs.FileInfo, err error) {
	log.Printf("a.statChildren(%#v) {\n", full)
	defer func() {
		log.Printf("a.statChildren(%#v) } -> (%p, %#v)\n", full, children, err)
	}()

	driverName, deviceName, remainder := a.pathParse(full)

	if driverName == "" {
		// root filesystem:
		children = make([]fs.FileInfo, 0, len(a.drivers))
		for _, d := range a.drivers {
			fi := &fileInfo{
				name:   d.Name,
				isDir:  true,
				driver: d,
			}
			children = append(children, fi)
		}
		return
	}

	// list one driver's devices:
	drv, ok := a.drivers[driverName]
	if !ok {
		err = fs.ErrNotExist
		return
	}

	// refresh device list for this driver:
	err = drv.refreshDevices()
	if err != nil {
		return
	}

	detected := drv.devices
	if deviceName == "" {
		// return the devices for the driver:
		children = make([]fs.FileInfo, 0, len(detected))
		for key := range detected {
			children = append(children, &fileInfo{
				name:      key,
				isDir:     true,
				driver:    drv,
				deviceKey: key,
			})
		}
		return
	}

	// find device:
	var device snes.AutoCloseableDevice
	device, ok = detected[deviceName]
	if !ok {
		err = fs.ErrNotExist
		return
	}

	// return actual files on the device:
	var entries []snes.DirEntry
	entries, err = device.ReadDirectory(ctx, remainder)
	if err != nil {
		return
	}

	children = make([]fs.FileInfo, 0, len(entries))
	for _, e := range entries {
		children = append(children, &fileInfo{
			name:      e.Name,
			isDir:     e.Type == sni.DirEntryType_Directory,
			driver:    drv,
			deviceKey: deviceName,
		})
	}

	return
}

func (a *AdapterFileSystem) stat(ctx context.Context, full cleanedPath) (stat *fileInfo, err error) {
	log.Printf("a.stat(%#v) {\n", full)
	defer func() {
		log.Printf("a.stat(%#v) } -> (%p, %#v)\n", full, stat, err)
	}()

	driverName, deviceName, remainder := a.pathParse(full)

	if driverName == "" {
		// root of filesystem:
		stat = &fileInfo{
			name:  "",
			isDir: true,
		}
		return
	}

	// look for driver:
	drv, ok := a.drivers[driverName]
	if !ok {
		err = fs.ErrNotExist
		return
	}

	if deviceName == "" {
		// return the stat for the driver:
		stat = &fileInfo{
			name:   drv.Name,
			isDir:  true,
			driver: drv,
		}
		return
	}

	// refresh device list for this driver:
	err = drv.refreshDevices()
	if err != nil {
		return
	}

	detected := drv.devices

	// find device:
	//var device snes.AutoCloseableDevice
	_, ok = detected[deviceName]
	if !ok {
		err = fs.ErrNotExist
		return
	}

	if remainder == "" {
		// stat for the device itself:
		stat = &fileInfo{
			name:      deviceName,
			isDir:     true,
			driver:    drv,
			deviceKey: deviceName,
		}
		return
	}

	// list from parent directory to find file:
	parent, file := path.Split(string(full))
	if len(file) == 0 {
		err = fs.ErrNotExist
		return
	}
	// remove trailing slash:
	if len(parent) > 0 && parent[len(parent)-1] == '/' {
		parent = parent[:len(parent)-1]
	}

	// look up parent directory listing to verify filename:
	var children []fs.FileInfo
	children, err = a.getStatChildren(ctx, cleanedPath(parent), strings.ToLower(parent))
	if err != nil {
		return
	}

	for _, e := range children {
		// found our file?
		if strings.EqualFold(e.Name(), file) {
			stat = e.(*fileInfo)
			stat.driver = drv
			stat.deviceKey = deviceName
			return
		}
	}

	//// exclude MacOS Finder metadata files:
	//_, file := filepath.Split(key)
	//if strings.HasPrefix(file, "._") {
	//	err = fs.ErrNotExist
	//	return
	//} else if file == ".DS_Store" {
	//	err = fs.ErrNotExist
	//	return
	//}

	err = fs.ErrNotExist
	return
}
