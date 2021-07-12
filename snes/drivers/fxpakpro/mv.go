package fxpakpro

import (
	"context"
	"fmt"
)

// mv does not allow moving files between folders; only renaming a file in an existing folder to a new filename
func (d *Device) mv(ctx context.Context, path, newFilename string) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpMV)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the current path to position 256:
	nameBytes := []byte(path)
	copy(sb[256:512], nameBytes)
	// copy in the new filename to position 8:
	copy(sb[8:256], newFilename)

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	// send command:
	err = sendSerial(d.f, 512, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	// read response:
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		return fmt.Errorf("mv: fxpakpro response packet does not contain USBA header")
	}
	if sb[4] != byte(OpRESPONSE) {
		_ = d.Close()
		return fmt.Errorf("mv: wrong opcode in response packet; got $%02x", sb[4])
	}
	if ec := sb[5]; ec != 0 {
		return fmt.Errorf("mv: %w", fxpakproError(ec))
	}

	return
}
