package webdav

import (
	"io/fs"
	"sni/snes"
)

type readable struct {
	fs   snes.DeviceFilesystem
	name string
}

func (f *readable) Close() error {
	//TODO implement me
	panic("implement me")
}

func (f *readable) Read(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (f *readable) Seek(offset int64, whence int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *readable) Readdir(count int) ([]fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f *readable) Stat() (fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f *readable) Write(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}
