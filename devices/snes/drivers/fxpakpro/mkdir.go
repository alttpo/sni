package fxpakpro

import (
	"context"
	"fmt"
)

func (d *Device) mkdir(ctx context.Context, path string) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpMKDIR)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(path)
	copy(sb[256:512], nameBytes)

	if shouldLock(ctx) {
		d.lock.Lock()
		defer d.lock.Unlock()
	}

	// send command:
	err = sendSerialChunked(d.f, 512, sb)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	// read response:
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		err = d.FatalError(err)
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		err = fmt.Errorf("mkdir: fxpakpro response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if sb[4] != byte(OpRESPONSE) {
		err = fmt.Errorf("mkdir: wrong opcode in response packet; got $%02x", sb[4])
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		err = fmt.Errorf("mkdir: %w", fxpakproError(ec))
		err = d.NonFatalError(err)
		return
	}

	return
}
