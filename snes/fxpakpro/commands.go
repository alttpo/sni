package fxpakpro

import (
	"encoding/binary"
	"fmt"
)

func (d *Device) put(space space, address uint32, data []byte) (err error) {
	sb := make([]byte, 512)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(space)
	sb[6] = byte(FlagNONE)

	// put the data size in:
	size := uint32(len(data))
	binary.BigEndian.PutUint32(sb[252:], size)

	// put the address in:
	binary.BigEndian.PutUint32(sb[256:], address)

	// send the data to the USB port:
	defer d.lock.Unlock()
	d.lock.Lock()

	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	dest := sb[0:]
	for len(data) > 0 {
		var n int
		for i := range dest {
			dest[i] = 0
		}
		n = copy(dest, data)
		data = data[n:]

		err = sendSerial(d.f, sb)
		if err != nil {
			_ = d.Close()
			return
		}
	}

	// await single response:
	err = recvSerial(d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}

	return
}

type vputChunk struct {
	addr uint32
	data []byte
}

func (d *Device) vput(space space, chunks []vputChunk) (err error) {
	if len(chunks) > 8 {
		return fmt.Errorf("VPUT cannot accept more than 8 chunks")
	}

	sb := make([]byte, 64)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpVPUT)
	sb[5] = byte(space)
	sb[6] = byte(FlagDATA64B | FlagNORESP)

	total := 0
	sp := sb[32:]
	for _, chunk := range chunks {
		if len(chunk.data) > 255 {
			return fmt.Errorf("VPUT chunk data size %d cannot exceed 255 bytes", len(chunk.data))
		}

		args := [4]byte{
			byte(len(chunk.data)),
			// big endian:
			byte((chunk.addr >> 16) & 0xFF),
			byte((chunk.addr >> 8) & 0xFF),
			byte((chunk.addr >> 0) & 0xFF),
		}
		copy(sp, args[:])
		sp = sp[4:]
		total += int(args[0])
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	// calculate expected number of packets:
	packets := total / 64
	remainder := total & 63
	if remainder > 0 {
		packets++
	}

	// concatenate all accompanying data together in one large slice:
	expected := packets * 64
	whole := make([]byte, expected)
	o := 0
	for _, chunk := range chunks {
		copy(whole[o:], chunk.data)
		o += len(chunk.data)
	}

	// send the expected number of 64-byte packets:
	err = sendSerial(d.f, whole)
	if err != nil {
		_ = d.Close()
		return
	}

	return
}
