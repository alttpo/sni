package snes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

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

// Header starts at $FFB0 to accommodate version 2 and 3 headers
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

	NativeVectors   NativeVectors
	EmulatedVectors EmulatedVectors
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

// WriteHeader always writes assuming it's starting at 0xFFB0
func (h *Header) WriteHeader(b *bytes.Buffer) (err error) {
	if err = writeBinaryStruct(b, h); err != nil {
		return
	}
	return
}

func (h *Header) Score(addr uint32) (score int) {
	// Reset vector must point to ROM at $00:8000-FFFF:
	if h.EmulatedVectors.RESET < 0x8000 {
		return 0
	}
	// Increase points for each other vector >= 0x8000:
	if h.NativeVectors.NMI >= 0x8000 {
		score++
	}
	if h.NativeVectors.BRK >= 0x8000 {
		score++
	}
	if h.NativeVectors.IRQ >= 0x8000 {
		score++
	}
	if h.NativeVectors.COP >= 0x8000 {
		score++
	}
	if h.NativeVectors.ABORT >= 0x8000 {
		score++
	}
	if h.EmulatedVectors.NMI >= 0x8000 {
		score++
	}
	if h.EmulatedVectors.IRQBRK >= 0x8000 {
		score++
	}
	if h.EmulatedVectors.COP >= 0x8000 {
		score++
	}
	if h.EmulatedVectors.ABORT >= 0x8000 {
		score++
	}

	// A valid checksum is very encouraging:
	if (uint32(h.CheckSum)+uint32(h.ComplementCheckSum) == 0xffff) && (h.CheckSum != 0) && (h.ComplementCheckSum != 0) {
		score += 8
	}
	if h.OldMakerCode == 0x33 {
		score += 2
	}

	// Valid ranges:
	if h.CartridgeType < 0x08 {
		score++
	}
	if h.ROMSize < 0x10 {
		score++
	}
	if h.RAMSize < 0x08 {
		score++
	}
	if h.DestinationCode < 0x0e {
		score++
	}

	// 0x20 is usually LoROM
	mapper := h.MapMode & ^uint8(0x10)
	if addr == 0x007fb0 && mapper == 0x20 {
		score += 2
	}
	// 0x21 is usually HiROM
	if addr == 0x00ffb0 && mapper == 0x21 {
		score += 2
	}
	// 0x22 is usually ExLoROM
	if addr == 0x007fb0 && mapper == 0x22 {
		score += 2
	}
	// 0x25 is usually ExHiROM
	if addr == 0x40ffb0 && mapper == 0x25 {
		score += 2
	}

	// TODO: could read RESET vector opcode to increase confidence

	return
}

func (h *Header) ROMSizeBytes() uint32 {
	return 1024 << h.ROMSize
}

func (h *Header) RAMSizeBytes() uint32 {
	return 1024 << h.RAMSize
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
