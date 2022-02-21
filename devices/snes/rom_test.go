package snes

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"testing"
)

func sampleROM() []byte {
	contents := make([]byte, 0x10000)
	hex.Decode(
		contents[0x7FB0:],
		[]byte("018d2401e2306bffffffffffffffffff"+
			"544845204c4547454e44204f46205a45"+
			"4c4441202020020a03010100f2500daf"+
			"ffffffff2c82ffff2c82c9800080d882"+
			"ffffffff2c822c822c822c820080d882"),
	)
	return contents
}

func TestNewROM(t *testing.T) {
	contents := sampleROM()
	gotR, err := NewROM("", contents)
	if err != nil {
		t.Fatal(err)
	}

	// check:
	if gotR.Header.NativeVectors.NMI != 0x80c9 {
		t.Fatal("NativeVectors.NMI")
	}
}

func TestROM_BusReader_Success(t *testing.T) {
	contents := sampleROM()
	rom, err := NewROM("", contents)
	if err != nil {
		t.Fatal(err)
	}

	r := rom.BusReader(0x00FFEA)
	p := uint16(0)
	err = binary.Read(r, binary.LittleEndian, &p)
	if err != nil {
		t.Fatal(err)
	}
	if p != 0x80c9 {
		t.Fatal("expected NMI vector at $FFEA")
	}
}

func TestROM_BusReader_Fail_Unmapped(t *testing.T) {
	contents := sampleROM()
	rom, err := NewROM("", contents)
	if err != nil {
		t.Fatal(err)
	}

	r := rom.BusReader(0x007FFF)
	p := uint16(0)
	err = binary.Read(r, binary.LittleEndian, &p)
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("expected fail with EOF but got: %v", err)
	}
}

func TestROM_BusReader_Fail_Boundary(t *testing.T) {
	contents := sampleROM()
	rom, err := NewROM("", contents)
	if err != nil {
		t.Fatal(err)
	}

	r := rom.BusReader(0x00FFFF)
	p := uint16(0)
	err = binary.Read(r, binary.LittleEndian, &p)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected fail with EOF but got: %v", err)
	}
}

func TestROM_BusWriter_Bank00_Success(t *testing.T) {
	contents := sampleROM()
	rom, err := NewROM("", contents)
	if err != nil {
		t.Fatal(err)
	}

	// write to bus writer:
	w := rom.BusWriter(0x00FFEA)
	p := uint16(0x80c8)
	err = binary.Write(w, binary.LittleEndian, &p)
	if err != nil {
		t.Fatal(err)
	}

	// read it from the rom.Contents:
	err = binary.Read(bytes.NewReader(rom.Contents[0x007FEA:]), binary.LittleEndian, &p)
	if err != nil {
		t.Fatal(err)
	}
	if p != 0x80c8 {
		t.Fatal("expected NMI vector at $FFEA")
	}

	// Also test ReadHeader for bank 00 access:
	err = rom.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	if rom.Header.NativeVectors.NMI != 0x80c8 {
		t.Fatal("ReadHeader should update NMI vector at $FFEA")
	}
}

func TestROM_BusWriter_Bank01_Success(t *testing.T) {
	contents := sampleROM()
	rom, err := NewROM("", contents)
	if err != nil {
		t.Fatal(err)
	}

	w := rom.BusWriter(0x018000)
	p := byte(0x80)
	err = binary.Write(w, binary.LittleEndian, &p)
	if err != nil {
		t.Fatal(err)
	}
	if p != 0x80 {
		t.Fatal("expected NMI vector at $FFEA")
	}
}
