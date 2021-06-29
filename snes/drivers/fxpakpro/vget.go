package fxpakpro

import (
	"fmt"
)

type vgetChunk struct {
	size   uint8
	addr   uint32
	target []byte
}

func (d *Device) vget(space space, chunks ...vgetChunk) (err error) {
	return d.vgetImpl(true, space, chunks...)
}

func (d *Device) vgetImpl(doLock bool, space space, chunks ...vgetChunk) (err error) {
	if len(chunks) > 8 {
		return fmt.Errorf("VGET cannot accept more than 8 chunks")
	}

	sb := make([]byte, 64)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpVGET)
	sb[5] = byte(space)
	sb[6] = byte(FlagDATA64B | FlagNORESP)

	total := 0
	sp := sb[32:]
	for _, chunk := range chunks {
		copy(sp, []byte{
			chunk.size,
			byte((chunk.addr >> 16) & 0xFF),
			byte((chunk.addr >> 8) & 0xFF),
			byte((chunk.addr >> 0) & 0xFF),
		})
		sp = sp[4:]
		total += int(chunk.size)
	}

	if doLock {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

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

	// read the expected number of 64-byte packets:
	expected := packets * 64
	rsp := make([]byte, expected)
	err = recvSerial(d.f, rsp, expected)
	if err != nil {
		_ = d.Close()
		return
	}

	// shrink down to exact size:
	rsp = rsp[0:total]
	start := 0
	for _, chunk := range chunks {
		end := start + int(chunk.size)
		// copy response data:
		copy(chunk.target, rsp[start:end])
		start = end
	}

	return
}
