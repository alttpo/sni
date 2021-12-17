package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sni/snes"
)

func (d *Device) getFile(ctx context.Context, path string, w io.Writer, sizeReceived snes.SizeReceivedFunc, progress snes.ProgressReportFunc) (received uint32, err error) {
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
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		err = d.FatalError(err)
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		received, err = 0, fmt.Errorf("getFile: response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if sb[4] != byte(OpRESPONSE) {
		_ = d.Close()
		received, err = 0, fmt.Errorf("getFile: wrong opcode in response packet; got $%02x", sb[4])
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		received, err = 0, fmt.Errorf("getFile: %w", fxpakproError(ec))
		err = d.NonFatalError(err)
		return
	}

	// read the size of the file:
	size := binary.BigEndian.Uint32(sb[252:256])
	if sizeReceived != nil {
		sizeReceived(size)
	}

	// read all remaining bytes in chunks of 512 bytes:
	received, err = recvSerialProgress(ctx, d.f, w, size, 512, progress)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	return
}
