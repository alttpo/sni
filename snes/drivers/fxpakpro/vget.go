package fxpakpro

import (
	"context"
	"fmt"
)

type vgetChunk struct {
	size   uint8
	addr   uint32
	target []byte
}

func (d *Device) vget(ctx context.Context, space space, chunks ...vgetChunk) (err error) {
	if len(chunks) > 8 {
		return fmt.Errorf("VGET cannot accept more than 8 chunks")
	}

	sb := make([]byte, 64)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpVGET)
	sb[5] = byte(space)
	sb[6] = byte(FlagDATA64B | FlagNORESP)

	total := uint32(0)
	sp := sb[32:]
	for _, chunk := range chunks {
		copy(sp, []byte{
			chunk.size,
			byte((chunk.addr >> 16) & 0xFF),
			byte((chunk.addr >> 8) & 0xFF),
			byte((chunk.addr >> 0) & 0xFF),
		})
		sp = sp[4:]
		total += uint32(chunk.size)
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

	// read the expected number of 64-byte packets:
	expected := packets * 64
	rsp := make([]byte, expected)
	err = recvSerial(ctx, d.f, rsp, expected)
	if err != nil {
		err = d.FatalError(err)
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
