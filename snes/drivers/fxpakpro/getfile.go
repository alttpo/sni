package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sni/snes"
)

func (d *Device) getFile(ctx context.Context, path string, w io.Writer, progress snes.ProgressReportFunc) (received uint64, err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpGET)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(path)
	copy(sb[256:512], nameBytes)

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
	err = recvSerial(d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		return 0, fmt.Errorf("getFile: response packet does not contain USBA header")
	}
	if ec := sb[5]; ec != 0 {
		return 0, fmt.Errorf("getFile: %w", fxpakproError(ec))
	}

	// read all remaining bytes in chunks of 512 bytes:
	size := uint64(binary.BigEndian.Uint32(sb[252:256]))
	received, err = recvSerialProgress(d.f, w, size, 512, progress)
	return
}
