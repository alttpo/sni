package webdav

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"sni/snes"
)

var ErrSeekForbidden = fmt.Errorf("seeking is forbidden")

type readable struct {
	a *AdapterFileSystem

	full      cleanedPath
	driver    *driverDevices
	device    snes.AutoCloseableDevice
	remainder string
	stat      *fileInfo
	children  []fs.FileInfo

	buf    []byte
	reader *bytes.Reader
}

func (f *readable) Close() error {
	log.Printf("readable(%p).Close()\n", f)
	f.buf = nil
	f.reader = nil
	return nil
}

func (f *readable) Read(p []byte) (n int, err error) {
	log.Printf("readable(%p).Read(%#v bytes)\n", f, len(p))

	ctx := context.Background()

	err = f.getFile(ctx)
	if err != nil {
		return
	}

	return f.reader.Read(p)
}

func (f *readable) getFile(ctx context.Context) (err error) {
	if f.buf == nil {
		// read entire file from device into memory:
		tmp := bytes.Buffer{}

		var m uint32
		m, err = f.device.GetFile(
			ctx,
			f.remainder,
			&tmp,
			func(size uint32) { tmp.Grow(int(size)) },
			nil)
		if err != nil {
			fatal := true
			if derr, ok := err.(snes.DeviceError); ok {
				fatal = derr.IsFatal()
			}
			if fatal {
				return
			} else {
				log.Printf("readable(%p).getFile(): %v\n", f, err)
				err = nil
			}
		}

		f.buf = tmp.Bytes()
		f.stat.size = m
	}

	if f.reader == nil {
		f.reader = bytes.NewReader(f.buf)
	}

	return
}

func (f *readable) Seek(offset int64, whence int) (n int64, err error) {
	log.Printf("readable(%p).Seek(%#v, %#v)\n", f, offset, whence)

	ctx := context.Background()

	err = f.getFile(ctx)
	if err != nil {
		return
	}

	return f.reader.Seek(offset, whence)
}

func (f *readable) Readdir(count int) (fis []fs.FileInfo, err error) {
	log.Printf("readable(%p).Readdir(%#v)\n", f, count)

	// NOTE: Readdir is stateful and should page through the dir entries each call and then return EOF
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
	log.Printf("readable(%p).Stat()\n", f)
	return f.stat, nil
}

func (f *readable) Write(p []byte) (n int, err error) {
	log.Printf("readable(%p).Write(%#v bytes)\n", f, len(p))
	return 0, fs.ErrInvalid
}
