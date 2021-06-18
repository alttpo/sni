package snes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type ROM struct {
	Name     string
	Contents []byte

	HeaderOffset    uint32
	Header          Header
	NativeVectors   NativeVectors
	EmulatedVectors EmulatedVectors
}

type Region uint8

const (
	RegionJapan Region = iota
	RegionNorthAmerica
	RegionEurope
	RegionSwedenScandinavia
	RegionFinland
	RegionDenmark
	RegionFrance
	RegionNetherlands
	RegionSpain
	RegionGermany
	RegionItaly
	RegionChina
	RegionIndonesia
	RegionKorea
	RegionGlobal
	RegionCanada
	RegionBrazil
	RegionAustralia
	RegionOther1
	RegionOther2
	RegionOther3
)

var RegionNames = map[Region]string{
	0x00: "Japan",
	0x01: "North America",
	0x02: "Europe",
	0x03: "Sweden/Scandinavia",
	0x04: "Finland",
	0x05: "Denmark",
	0x06: "France",
	0x07: "Netherlands",
	0x08: "Spain",
	0x09: "Germany",
	0x0A: "Italy",
	0x0B: "China",
	0x0C: "Indonesia",
	0x0D: "Korea",
	0x0E: "Global (?)",
	0x0F: "Canada",
	0x10: "Brazil",
	0x11: "Australia",
	0x12: "Other (1)",
	0x13: "Other (2)",
	0x14: "Other (3)",
}

// $FFB0
type Header struct {
	version int // 1, 2, or 3

	// ver 2&3 header:
	MakerCode        uint16  `rom:"FFB0"`
	GameCode         uint32  `rom:"FFB2"`
	Fixed1           [6]byte //`rom:"FFB6"`
	FlashSize        byte    `rom:"FFBC"`
	ExpansionRAMSize byte    `rom:"FFBD"`
	SpecialVersion   byte    `rom:"FFBE"`
	CoCPUType        byte    `rom:"FFBF"`
	// ver 1 header:
	Title              [21]byte `rom:"FFC0"`
	MapMode            byte     `rom:"FFD5"`
	CartridgeType      byte     `rom:"FFD6"`
	ROMSize            byte     `rom:"FFD7"`
	RAMSize            byte     `rom:"FFD8"`
	DestinationCode    Region   `rom:"FFD9"`
	OldMakerCode       byte     `rom:"FFDA"` // = $33 to indicate ver 3 header
	MaskROMVersion     byte     `rom:"FFDB"`
	ComplementCheckSum uint16   `rom:"FFDC"`
	CheckSum           uint16   `rom:"FFDE"`
}

func (h *Header) HeaderVersion() int { return h.version }

// ReadHeader parses a ROM header starting from FFB0 up to FFE0
func (h *Header) ReadHeader(b *bytes.Reader) (err error) {
	// Read SNES header:
	if err = readBinaryStruct(b, h); err != nil {
		return
	}

	if h.OldMakerCode == 0x33 {
		h.version = 3
	} else if h.Title[20] == 0x00 {
		h.version = 2
	} else {
		h.version = 1
		// Zero-out all the version 2&3 fields:
		h.MakerCode = 0
		h.GameCode = 0
		h.Fixed1 = [6]byte{}
		h.FlashSize = 0
		h.ExpansionRAMSize = 0
		h.SpecialVersion = 0
		h.CoCPUType = 0
	}

	return
}

func (h *Header) WriteHeader(b *bytes.Buffer) (err error) {
	err = writeBinaryStruct(b, h)
	return
}

type NativeVectors struct {
	Unused1 [4]byte //`rom:"FFE0"`
	COP     uint16  `rom:"FFE4"`
	BRK     uint16  `rom:"FFE6"`
	ABORT   uint16  `rom:"FFE8"`
	NMI     uint16  `rom:"FFEA"`
	Unused2 uint16  //`rom:"FFEC"`
	IRQ     uint16  `rom:"FFEE"`
}

type EmulatedVectors struct {
	Unused1 [4]byte //`rom:"FFF0"`
	COP     uint16  `rom:"FFF4"`
	Unused2 uint16  //`rom:"FFF6"`
	ABORT   uint16  `rom:"FFF8"`
	NMI     uint16  `rom:"FFFA"`
	RESET   uint16  `rom:"FFFC"`
	IRQBRK  uint16  `rom:"FFFE"`
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
	if err = readBinaryStruct(b, &r.Header); err != nil {
		return
	}

	if r.Header.OldMakerCode == 0x33 {
		r.Header.version = 3
	} else if r.Header.Title[20] == 0x00 {
		r.Header.version = 2
	} else {
		r.Header.version = 1
		// Zero-out all the version 2&3 fields:
		r.Header.MakerCode = 0
		r.Header.GameCode = 0
		r.Header.Fixed1 = [6]byte{}
		r.Header.FlashSize = 0
		r.Header.ExpansionRAMSize = 0
		r.Header.SpecialVersion = 0
		r.Header.CoCPUType = 0
	}

	if err = readBinaryStruct(b, &r.NativeVectors); err != nil {
		return
	}
	if err = readBinaryStruct(b, &r.EmulatedVectors); err != nil {
		return
	}
	return
}

func (r *ROM) WriteHeader() (err error) {
	var b = &bytes.Buffer{}
	if err = writeBinaryStruct(b, &r.Header); err != nil {
		return
	}
	if err = writeBinaryStruct(b, &r.NativeVectors); err != nil {
		return
	}
	if err = writeBinaryStruct(b, &r.EmulatedVectors); err != nil {
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

func readBinaryStruct(b *bytes.Reader, into interface{}) (err error) {
	hv := reflect.ValueOf(into).Elem()
	for i := 0; i < hv.NumField(); i++ {
		f := hv.Field(i)
		// skip unexported fields:
		if !f.CanInterface() {
			continue
		}

		var p interface{}

		if !f.CanAddr() {
			panic(fmt.Errorf("error handling struct field %s of type %s; cannot take address of field", hv.Type().Field(i).Name, hv.Type().Name()))
		}

		p = f.Addr().Interface()
		err = binary.Read(b, binary.LittleEndian, p)
		if err != nil {
			return fmt.Errorf("error reading struct field %s of type %s: %w", hv.Type().Field(i).Name, hv.Type().Name(), err)
		}
		//fmt.Printf("%s: %v\n", reflect.TypeOf(r.Header).Field(i).Name, f.Interface())
	}
	return
}

func writeBinaryStruct(w io.Writer, from interface{}) (err error) {
	hv := reflect.ValueOf(from).Elem()
	for i := 0; i < hv.NumField(); i++ {
		f := hv.Field(i)
		// skip unexported fields:
		if !f.CanInterface() {
			continue
		}

		if !f.CanAddr() {
			panic(fmt.Errorf("error handling struct field %s of type %s; cannot take address of field", hv.Type().Field(i).Name, hv.Type().Name()))
		}

		var p interface{}
		p = f.Addr().Interface()
		err = binary.Write(w, binary.LittleEndian, p)
		if err != nil {
			return fmt.Errorf("error writing struct field %s of type %s: %w", hv.Type().Field(i).Name, hv.Type().Name(), err)
		}
		//fmt.Printf("%s: %v\n", reflect.TypeOf(r.Header).Field(i).Name, f.Interface())
	}
	return
}

func (r *ROM) ROMSize() uint32 {
	return 1024 << r.Header.ROMSize
}

func (r *ROM) RAMSize() uint32 {
	return 1024 << r.Header.RAMSize
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
