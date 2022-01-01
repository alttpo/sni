package webdav

import (
	"context"
	dav "golang.org/x/net/webdav"
	"io/fs"
	"log"
	"os"
)

// AdapterFileSystem adapts the underlying simplistic snes.DeviceFilesystem interface to the more flexible
// dav.FileSystem interface. The major limitation of the snes.DeviceFilesystem is the lack of random access
// on files. The current sd2snes/fxpakpro firmware (v1.11.0) is limited to transferring files (GET/PUT)
// sequentially from start to end without interruption. Thus, the dav.File instance returned from OpenFile
// must ensure that reading or writing is done sequentially and must not allow seeking after a read or write
// has started.
type AdapterFileSystem struct {
	//fs snes.DeviceFilesystem
}

func (a *AdapterFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Printf("mkdir(%#v, %#v)\n", name, perm)

	_ = perm
	//return a.fs.MakeDirectory(ctx, name)
	return nil
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
		var stat os.FileInfo
		stat, err = a.Stat(ctx, name)
		if err != nil {
			return
		}

		f = NewReadable(a, name, stat)
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

	if name == "" {
		stat = &fileInfo{
			name:  "",
			isDir: true,
		}
		return
	}

	return nil, fs.ErrNotExist
}
