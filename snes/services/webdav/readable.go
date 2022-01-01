package webdav

import (
	"fmt"
	"io/fs"
	"log"
	"os"
)

var ErrSeekForbidden error = fmt.Errorf("seeking is forbidden")

type readable struct {
	a *AdapterFileSystem
	//fs   snes.DeviceFilesystem

	name string
	stat os.FileInfo

	children []fs.FileInfo
}

func NewReadable(a *AdapterFileSystem, name string, stat os.FileInfo) (f *readable) {
	f = &readable{
		a:    a,
		name: name,
		stat: stat,
	}
	return
}

func (f *readable) Close() error {
	log.Printf("%p.close()\n", f)
	return nil
}

func (f *readable) Read(p []byte) (n int, err error) {
	log.Printf("%p.read(%p)\n", f, p)
	err = os.ErrInvalid
	return
}

func (f *readable) Seek(offset int64, whence int) (n int64, err error) {
	log.Printf("%p.seek(%#v, %#v)\n", f, offset, whence)
	return 0, ErrSeekForbidden
}

func (f *readable) Readdir(count int) (fis []fs.FileInfo, err error) {
	log.Printf("%p.readdir(%#v)\n", f, count)

	// TODO: Readdir is stateful and should page through the dir entries each call and then return EOF
	return f.children[0:count], nil
}

func (f *readable) Stat() (fi fs.FileInfo, err error) {
	log.Printf("%p.stat()\n", f)
	return f.stat, nil
}

func (f *readable) Write(p []byte) (n int, err error) {
	log.Printf("%p.write(%p)\n", f, p)
	return 0, fs.ErrInvalid
}
