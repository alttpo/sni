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
	"strconv"
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

	drivers map[string]snes.NamedDriver
}

func NewAdapterFileSystem() (a *AdapterFileSystem) {
	a = &AdapterFileSystem{
		statsC:    cache.New(time.Second, time.Second*15),
		childrenC: cache.New(time.Second, time.Second*15),
		drivers:   make(map[string]snes.NamedDriver),
	}

	all := snes.Drivers()
	for _, d := range all {
		// make sure the driver has filesystem caps:
		ok, _ := d.Driver.HasCapabilities(sni.DeviceCapability_ReadDirectory)
		if !ok {
			continue
		}

		key := strings.ToLower(d.Name)
		a.drivers[key] = d
	}

	return
}

func (a *AdapterFileSystem) getStat(ctx context.Context, name string) (stat *fileInfo, err error) {
	full, key := a.pathClean(name)

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

func (a *AdapterFileSystem) getStatChildren(ctx context.Context, name string) (children []fs.FileInfo, err error) {
	full, key := a.pathClean(name)

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
		fullPath := path.Join(name, child.Name())
		fullPathKey := strings.ToLower(fullPath)
		a.statsC.Set(fullPathKey, child, cache.DefaultExpiration)
	}
	return
}

func (a *AdapterFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Printf("a.Mkdir(%#v, %#v)\n", name, perm)

	_ = perm
	//return a.fs.MakeDirectory(ctx, name)
	return fs.ErrInvalid
}

func (a *AdapterFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (f dav.File, err error) {
	log.Printf("a.OpenFile(%#v, %#v, %#v) {\n", name, flag, perm)
	defer func() {
		log.Printf("a.OpenFile(%#v, %#v, %#v) } -> (%#v, %#v)\n", name, flag, perm, f, err)
	}()
	if flag&os.O_RDWR != 0 {
		// cannot open files in both read/write mode:
		err = os.ErrInvalid
		return
	}
	if flag&os.O_WRONLY != 0 {
		f = &writeable{name: name}
		return
	} else {
		// readable open:
		var stat *fileInfo
		stat, err = a.getStat(ctx, name)
		if err != nil {
			return
		}

		var children []fs.FileInfo
		if stat.IsDir() {
			children, err = a.getStatChildren(ctx, name)
			if err != nil {
				return
			}
		}

		f = NewReadable(a, name, stat, children)
		return
	}
}

func (a *AdapterFileSystem) RemoveAll(ctx context.Context, name string) error {
	log.Printf("a.RemoveAll(%#v)\n", name)
	return nil
}

func (a *AdapterFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	log.Printf("a.Rename(%#v, %#v)\n", oldName, newName)
	return nil
}

func (a *AdapterFileSystem) Stat(ctx context.Context, name string) (stat os.FileInfo, err error) {
	log.Printf("a.Stat(%#v)\n", name)

	return a.getStat(ctx, name)
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
		log.Printf("a.statChildren(%#v) } -> (%#v, %#v)\n", full, children, err)
	}()

	driverName, deviceName, remainder := a.pathParse(full)

	if driverName == "" {
		// root filesystem:
		children = make([]fs.FileInfo, 0, len(a.drivers))
		for _, d := range a.drivers {
			fi := &fileInfo{
				name:  d.Name,
				isDir: true,
			}
			children = append(children, fi)
		}
		return
	}

	// list one driver's devices:
	driver, ok := a.drivers[driverName]
	if !ok {
		err = fs.ErrNotExist
		return
	}

	var detected []snes.DeviceDescriptor
	detected, err = driver.Driver.Detect()
	if err != nil {
		return
	}

	if deviceName == "" {
		// return the devices for the driver:
		children = make([]fs.FileInfo, 0, len(detected))
		for i := range detected {
			children = append(children, &fileInfo{
				// use the device index and hope it's stable:
				// todo: alternatively sanitize device URL as a filename
				name:  strconv.Itoa(i),
				isDir: true,
			})
		}
		return
	}

	// find device:
	var deviceDesc *snes.DeviceDescriptor
	for i, d := range detected {
		// find by index:
		if deviceName == strconv.Itoa(i) {
			deviceDesc = &d
			break
		}
	}
	if deviceDesc == nil {
		err = fs.ErrNotExist
		return
	}

	var device snes.AutoCloseableDevice
	_, device, err = snes.DeviceByUri(&deviceDesc.Uri)
	if err != nil {
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
			name:  e.Name,
			isDir: e.Type == sni.DirEntryType_Directory,
		})
	}

	return
}

func (a *AdapterFileSystem) stat(ctx context.Context, full cleanedPath) (stat *fileInfo, err error) {
	log.Printf("a.stat(%#v) {\n", full)
	defer func() {
		log.Printf("a.stat(%#v) } -> (%#v, %#v)\n", full, stat, err)
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
	driver, ok := a.drivers[driverName]
	if !ok {
		err = fs.ErrNotExist
		return
	}

	if deviceName == "" {
		// return the stat for the driver:
		stat = &fileInfo{
			name:  driver.Name,
			isDir: true,
		}
		return
	}

	var detected []snes.DeviceDescriptor
	detected, err = driver.Driver.Detect()
	if err != nil {
		return
	}

	// find device:
	var deviceDesc *snes.DeviceDescriptor
	for i, d := range detected {
		// find by index:
		if deviceName == strconv.Itoa(i) {
			deviceDesc = &d
			break
		}
	}
	if deviceDesc == nil {
		err = fs.ErrNotExist
		return
	}

	if remainder == "" {
		// stat for the device itself:
		stat = &fileInfo{
			name:  deviceName,
			isDir: true,
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
	children, err = a.getStatChildren(ctx, parent)
	if err != nil {
		return
	}

	for _, e := range children {
		// found our file?
		if strings.EqualFold(e.Name(), file) {
			stat = e.(*fileInfo)
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
