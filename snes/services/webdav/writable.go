package webdav

import (
	"bytes"
	"context"
	"io/fs"
	"log"
	"os"
	"path"
	"sni/snes"
)

type writeable struct {
	a *AdapterFileSystem

	full      cleanedPath
	remainder string
	stat      *fileInfo

	driver *driverDevices
	device snes.AutoCloseableDevice

	buf []byte
}

func (f *writeable) Close() (err error) {
	log.Printf("%p.close()\n", f)

	// put the file into the device:
	var n uint32
	n, err = f.device.PutFile(
		context.Background(),
		f.remainder,
		uint32(len(f.buf)),
		bytes.NewReader(f.buf),
		nil)
	if err != nil {
		return
	}
	_ = n

	// invalidate cache:
	full := string(f.full)
	f.a.statsC.Delete(full)
	parent, _ := path.Split(full)
	f.a.childrenC.Delete(parent)

	return
}

func (f *writeable) Readdir(count int) (fis []fs.FileInfo, err error) {
	log.Printf("%p.readdir(%#v)\n", f, count)
	return nil, fs.ErrInvalid
}

func (f *writeable) Stat() (fi fs.FileInfo, err error) {
	log.Printf("%p.stat()\n", f)
	return f.stat, nil
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

func (f *writeable) Write(p []byte) (n int, err error) {
	log.Printf("%p.write(%p)\n", f, p)
	f.buf = append(f.buf, p...)
	return len(p), nil
}
