package snes

import (
	"bytes"
	"fmt"
	"io"
)

type ROM struct {
	Name     string
	Contents []byte

	HeaderOffset uint32
	Header       Header
}

func NewROM(name string, contents []byte) (r *ROM, err error) {
	if len(contents) < 0x8000 {
		return nil, fmt.Errorf("ROM file not big enough to contain SNES header")
	}

	headerOffset := uint32(0x007FB0)

	r = &ROM{
		Name:         name,
		Contents:     contents,
		HeaderOffset: headerOffset,
	}

	err = r.ReadHeader()
	return
}

func (r *ROM) ReadHeader() (err error) {
	// Read SNES header:
	b := bytes.NewReader(r.Contents[r.HeaderOffset : r.HeaderOffset+0x50])
	err = r.Header.ReadHeader(b)
	return
}

func (r *ROM) WriteHeader() (err error) {
	var b = &bytes.Buffer{}
	if err = r.Header.WriteHeader(b); err != nil {
		return
	}

	if r.Header.version <= 1 {
		// overwrite FFC0 if version 1 (leave FFB0-BF untouched):
		copy(r.Contents[r.HeaderOffset+0x10:r.HeaderOffset+0x50], b.Bytes()[0x10:])
	} else {
		// overwrite FFB0 if version 2 or 3:
		copy(r.Contents[r.HeaderOffset:r.HeaderOffset+0x50], b.Bytes())
	}
	return
}

type alwaysError struct{}

func (alwaysError) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (alwaysError) Write(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

var alwaysErrorInstance = &alwaysError{}

func (r *ROM) BusReader(busAddr uint32) io.Reader {
	page := busAddr & 0xFFFF
	if page < 0x8000 {
		return alwaysErrorInstance
	}

	// Return a reader over the ROM contents up to the next bank to prevent accidental overflow:
	bank := busAddr >> 16
	pcStart := (bank << 15) | (page - 0x8000)
	pcEnd := (bank << 15) | 0x7FFF
	return bytes.NewReader(r.Contents[pcStart:pcEnd])
}

type busWriter struct {
	r       *ROM
	busAddr uint32
	start   uint32
	end     uint32
	o       uint32
}

func (w *busWriter) Write(p []byte) (n int, err error) {
	if uint32(len(p)) >= w.o+w.end {
		err = io.ErrUnexpectedEOF
		return
	}

	n = copy(w.r.Contents[w.o+w.start:w.end], p)
	w.o += uint32(n)

	return
}

func (r *ROM) BusWriter(busAddr uint32) io.Writer {
	page := busAddr & 0xFFFF
	if page < 0x8000 {
		return alwaysErrorInstance
	}

	// Return a reader over the ROM contents up to the next bank to prevent accidental overflow:
	bank := busAddr >> 16
	pcStart := (bank << 15) | (page - 0x8000)
	pcEnd := (bank << 15) | 0x7FFF
	return &busWriter{r, busAddr, pcStart, pcEnd, 0}
}
