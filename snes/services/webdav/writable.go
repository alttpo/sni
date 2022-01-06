package webdav

import (
	"bytes"
	"context"
	"io/fs"
	"log"
	"os"
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
	log.Printf("writable(%p).Close()\n", f)

	if len(f.buf) > 0 {
		// put the file into the device:
		var n uint32
		n, err = f.device.PutFile(
			context.Background(),
			f.remainder,
			uint32(len(f.buf)),
			bytes.NewReader(f.buf),
			nil)
		if err != nil {
			fatal := true
			if derr, ok := err.(snes.DeviceError); ok {
				fatal = derr.IsFatal()
			}
			if fatal {
				return
			} else {
				log.Printf("writable(%p).Close(): %v\n", f, err)
				err = nil
			}
		}
		_ = n
	} else {

	}

	// invalidate cache:
	f.a.invalidateStat(f.full)
	return
}

func (f *writeable) Readdir(count int) (fis []fs.FileInfo, err error) {
	log.Printf("writable(%p).Readdir(%#v)\n", f, count)
	return nil, fs.ErrInvalid
}

func (f *writeable) Stat() (fi fs.FileInfo, err error) {
	log.Printf("writable(%p).Stat()\n", f)
	return f.stat, nil
}

func (f *writeable) Read(p []byte) (n int, err error) {
	log.Printf("writable(%p).Read(%#v bytes)\n", f, len(p))
	err = os.ErrInvalid
	return
}

func (f *writeable) Seek(offset int64, whence int) (n int64, err error) {
	log.Printf("writable(%p).Seek(%#v, %#v)\n", f, offset, whence)
	return 0, ErrSeekForbidden
}

func (f *writeable) Write(p []byte) (n int, err error) {
	log.Printf("writable(%p).Write(%#v bytes)\n", f, len(p))
	f.buf = append(f.buf, p...)
	return len(p), nil
}
