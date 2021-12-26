package webdav

import (
	"io/fs"
	"time"
)

type fileInfo struct {
	name  string
	isDir bool
}

func (f *fileInfo) Name() string {
	//TODO implement me
	panic("implement me")
}

func (f *fileInfo) Size() int64 {
	//TODO implement me
	panic("implement me")
}

func (f *fileInfo) Mode() fs.FileMode {
	//TODO implement me
	panic("implement me")
}

func (f *fileInfo) ModTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (f *fileInfo) IsDir() bool {
	//TODO implement me
	panic("implement me")
}

func (f *fileInfo) Sys() interface{} {
	//TODO implement me
	panic("implement me")
}
