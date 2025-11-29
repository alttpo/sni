package fxpakpro

import (
	"context"
	"encoding/hex"
	"fmt"
	"runtime/trace"
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

	ctx, task := trace.NewTask(ctx, "fxpakpro:vget")
	defer task.End()

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
		d.lock.Lock()
		defer d.lock.Unlock()
	}

	if trace.IsEnabled() {
		trace.Log(ctx, "req", hex.Dump(sb))
	}

	err = sendSerial(ctx, d.f, sb)
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

	if trace.IsEnabled() {
		trace.Log(ctx, "rsp", hex.Dump(rsp))
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
