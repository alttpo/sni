package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sni/snes"
)

type putFileRequest struct {
	path   string
	rom    []byte
	report snes.ProgressReportFunc
}

func (d *Device) putFile(ctx context.Context, path string, size uint32, r io.Reader, progress snes.ProgressReportFunc) (n uint32, err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(path)
	copy(sb[256:512], nameBytes)

	// size of ROM contents:
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

	// read response:
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		return size, fmt.Errorf("putfile: response packet does not contain USBA header")
	}
	if sb[4] != byte(OpRESPONSE) {
		_ = d.Close()
		return size, fmt.Errorf("putfile: wrong opcode in response packet; got $%02x", sb[4])
	}
	if ec := sb[5]; ec != 0 {
		return size, fmt.Errorf("putfile: %w", fxpakproError(ec))
	}

	// send data:
	n, err = sendSerialProgress(d.f, 512, size, r, progress)
	if err != nil {
		_ = d.Close()
		return
	}

	return
}
