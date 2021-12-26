package webdav

import (
	"io/fs"
	"os"
	"sni/snes"
)

type writeable struct {
	fs   snes.DeviceFilesystem
	name string
}

func (f *writeable) Close() error {
	//TODO implement me
	panic("implement me")
}

func (f *writeable) Read(p []byte) (n int, err error) {
	err = os.ErrInvalid
	return
}

func (f *writeable) Seek(offset int64, whence int) (n int64, err error) {
	err = os.ErrInvalid
	return
}

func (f *writeable) Readdir(count int) (fis []fs.FileInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (f *writeable) Stat() (fi fs.FileInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (f *writeable) Write(p []byte) (n int, err error) {
	panic("TODO")
}
