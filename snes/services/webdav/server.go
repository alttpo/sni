package webdav

import (
	"context"
	dav "golang.org/x/net/webdav"
	"os"
	"sni/snes"
)

// AdapterFileSystem adapts the underlying simplistic snes.DeviceFilesystem interface to the more flexible
// dav.FileSystem interface. The major limitation of the snes.DeviceFilesystem is the lack of random access
// on files. The current sd2snes/fxpakpro firmware (v1.11.0) is limited to transferring files (GET/PUT)
// sequentially from start to end without interruption. Thus, the dav.File instance returned from OpenFile
// must ensure that reading or writing is done sequentially and must not allow seeking after a read or write
// has started.
type AdapterFileSystem struct {
	fs snes.DeviceFilesystem
}

func (a *AdapterFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	_ = perm
	return a.fs.MakeDirectory(ctx, name)
}

func (a *AdapterFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (f dav.File, err error) {
	if flag&os.O_RDWR != 0 {
		// cannot open files in both read/write mode:
		err = os.ErrInvalid
		return
	}
	if flag&os.O_WRONLY != 0 {
		f = &writeable{fs: a.fs, name: name}
		return
	} else {
		f = &readable{fs: a.fs, name: name}
		return
	}
}

func (a *AdapterFileSystem) RemoveAll(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}

func (a *AdapterFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return a.fs.RenameFile(ctx, oldName, newName)
}

func (a *AdapterFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return nil, nil
}
