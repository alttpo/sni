package fxpakpro

import (
	"context"
	"fmt"
)

type vputChunk struct {
	addr uint32
	data []byte
}

func (d *Device) vput(ctx context.Context, space space, chunks ...vputChunk) (err error) {
	if len(chunks) > 8 {
		return fmt.Errorf("VPUT cannot accept more than 8 chunks")
	}

	sb := make([]byte, 64)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
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

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	err = sendSerial(d.f, 64, sb)
	if err != nil {
		err = d.FatalError(err)
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
	err = sendSerial(d.f, 64, whole)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	return
}
