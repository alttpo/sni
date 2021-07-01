package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"sni/snes"
)

type putFileRequest struct {
	path   string
	rom    []byte
	report snes.ProgressReportFunc
}

func (d *Device) putFile(ctx context.Context, req putFileRequest) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(req.path)
	copy(sb[256:512], nameBytes)

	// size of ROM contents:
	size := uint32(len(req.rom))
	binary.BigEndian.PutUint32(sb[252:], size)

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

	// send data:
	err = sendSerialProgress(d.f, 512, req.rom, req.report)
	if err != nil {
		_ = d.Close()
		return
	}

	// read response:
	err = recvSerial(d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		return fmt.Errorf("putfile: response packet does not contain USBA header")
	}
	if ec := sb[5]; ec != 0 {
		return fmt.Errorf("putfile: %w", fxpakproError(ec))
	}

	return
}
