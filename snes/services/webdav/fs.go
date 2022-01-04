package webdav

import (
	"context"
	dav "golang.org/x/net/webdav"
	"io/fs"
	"log"
	"os"
	"path"
	"sni/protos/sni"
	"sni/snes"
	"strconv"
	"strings"
	"sync"
)

// AdapterFileSystem adapts the underlying simplistic snes.DeviceFilesystem interface to the more flexible
// dav.FileSystem interface. The major limitation of the snes.DeviceFilesystem is the lack of random access
// on files. The current sd2snes/fxpakpro firmware (v1.11.0) is limited to transferring files (GET/PUT)
// sequentially from start to end without interruption. Thus, the dav.File instance returned from OpenFile
// must ensure that reading or writing is done sequentially and must not allow seeking after a read or write
// has started.
type AdapterFileSystem struct {
	mu sync.Mutex

	stats    map[string]*fileInfo
	children map[string][]fs.FileInfo

	drivers map[string]snes.NamedDriver
}

func NewAdapterFileSystem() (a *AdapterFileSystem) {
	a = &AdapterFileSystem{
		stats:    make(map[string]*fileInfo),
		children: make(map[string][]fs.FileInfo),
		drivers:  make(map[string]snes.NamedDriver),
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

func (a *AdapterFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Printf("mkdir(%#v, %#v)\n", name, perm)

	_ = perm
	//return a.fs.MakeDirectory(ctx, name)
	return fs.ErrInvalid
}

func (a *AdapterFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (f dav.File, err error) {
	defer func() {
		log.Printf("%p = openFile(%#v, %#v, %#v)\n", f, name, flag, perm)
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
		stat, err = a.stat(ctx, name)
		if err != nil {
			return
		}

		var children []fs.FileInfo
		if stat.IsDir() {
			children, err = a.statChildren(ctx, name)
			if err != nil {
				return
			}
		}

		f = NewReadable(a, name, stat, children)
		return
	}
}

func (a *AdapterFileSystem) RemoveAll(ctx context.Context, name string) error {
	log.Printf("removeAll(%#v)\n", name)
	panic("implement me")
}

func (a *AdapterFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	log.Printf("rename(%#v, %#v)\n", oldName, newName)
	return nil
	//return a.fs.RenameFile(ctx, oldName, newName)
}

func (a *AdapterFileSystem) Stat(ctx context.Context, name string) (stat os.FileInfo, err error) {
	log.Printf("stat(%#v)\n", name)

	stat, err = a.stat(ctx, name)
	return
}

func (a *AdapterFileSystem) pathParse(name string) (full, key, driverName, deviceName, remainder string) {
	// make sure we at least start with a /:
	if name == "" {
		name = "/"
	}

	full = path.Clean(name)
	key = strings.ToLower(full)

	parts := strings.Split(full[1:], "/")
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

func (a *AdapterFileSystem) statChildren(ctx context.Context, name string) (children []fs.FileInfo, err error) {
	defer a.mu.Unlock()
	a.mu.Lock()

	full, key, driverName, deviceName, remainder := a.pathParse(name)
	_ = full
	_ = remainder

	var ok bool
	children, ok = a.children[key]
	if ok {
		delete(a.children, key)
		return
	}

	if driverName == "" {
		// root filesystem:
		children = make([]fs.FileInfo, 0, len(a.drivers))
		for _, d := range a.drivers {
			fi := &fileInfo{
				name:  d.Name,
				isDir: true,
			}
			a.stats[fi.name] = fi
			children = append(children, fi)
		}
		a.children[""] = children
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

func (a *AdapterFileSystem) stat(ctx context.Context, name string) (stat *fileInfo, err error) {
	defer a.mu.Unlock()
	a.mu.Lock()

	full, key, driverName, deviceName, remainder := a.pathParse(name)
	_ = full
	_ = remainder

	var ok bool
	stat, ok = a.stats[key]
	if ok {
		delete(a.stats, key)
		return
	}

	if driverName == "" {
		// root of filesystem:
		stat = &fileInfo{
			name:  "",
			isDir: true,
		}
		a.stats[""] = stat
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

	var device snes.AutoCloseableDevice
	_, device, err = snes.DeviceByUri(&deviceDesc.Uri)
	if err != nil {
		return
	}

	// list from parent directory to find file:
	parent, file := path.Split(remainder)

	// return actual files on the device:
	var entries []snes.DirEntry
	entries, err = device.ReadDirectory(ctx, parent)
	if err != nil {
		return
	}

	for _, e := range entries {
		// found our file?
		if strings.EqualFold(e.Name, file) {
			stat = &fileInfo{
				name:  e.Name,
				isDir: e.Type == sni.DirEntryType_Directory,
			}
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
