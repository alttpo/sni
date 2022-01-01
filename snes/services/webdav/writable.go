package webdav

import (
	"io/fs"
	"log"
	"os"
)

type writeable struct {
	//fs   snes.DeviceFilesystem
	name string
	stat fileInfo
}

func (f *writeable) Close() error {
	log.Printf("%p.close()\n", f)
	return nil
}

func (f *writeable) Read(p []byte) (n int, err error) {
	log.Printf("%p.read(%p)\n", f, p)
	err = os.ErrInvalid
	return
}

func (f *writeable) Seek(offset int64, whence int) (n int64, err error) {
	log.Printf("%p.seek(%#v, %#v)\n", f, offset, whence)
	return 0, ErrSeekForbidden
}

func (f *writeable) Readdir(count int) (fis []fs.FileInfo, err error) {
	log.Printf("%p.readdir(%#v)\n", f, count)
	return nil, fs.ErrInvalid
}

func (f *writeable) Stat() (fi fs.FileInfo, err error) {
	log.Printf("%p.stat()\n", f)
	return &f.stat, nil
}

func (f *writeable) Write(p []byte) (n int, err error) {
	log.Printf("%p.write(%p)\n", f, p)
	return 0, nil
}
