package webdav

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
)

var ErrSeekForbidden = fmt.Errorf("seeking is forbidden")

type readable struct {
	a *AdapterFileSystem
	//fs   snes.DeviceFilesystem

	name     string
	stat     *fileInfo
	children []fs.FileInfo
}

func NewReadable(a *AdapterFileSystem, name string, stat *fileInfo) (f *readable) {
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

	// NOTE: Readdir is stateful and should page through the dir entries each call and then return EOF
	if f.children == nil {

	}

	if len(f.children) == 0 {
		fis = f.children[0:0]
		err = io.EOF
		return
	}

	if count <= 0 {
		fis = f.children[:]
		f.children = f.children[0:0]
	} else {
		if count >= len(f.children) {
			count = len(f.children)
		}
		fis = f.children[0:count]
		f.children = f.children[count:0]
	}

	return
}

func (f *readable) Stat() (fi fs.FileInfo, err error) {
	log.Printf("%p.stat()\n", f)
	return f.stat, nil
}

func (f *readable) Write(p []byte) (n int, err error) {
	log.Printf("%p.write(%p)\n", f, p)
	return 0, fs.ErrInvalid
}
