package webdav

import (
	"io/fs"
	"time"
)

type fileInfo struct {
	name  string
	isDir bool

	driver    *driverDevices
	deviceKey string
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return 0
}

func (f *fileInfo) Mode() (mode fs.FileMode) {
	mode = 0
	if f.isDir {
		mode |= fs.ModeDir
	}
	return
}

func (f *fileInfo) ModTime() time.Time {
	return time.Time{}
}

func (f *fileInfo) IsDir() bool {
	return f.isDir
}

func (f *fileInfo) Sys() interface{} {
	return nil
}
