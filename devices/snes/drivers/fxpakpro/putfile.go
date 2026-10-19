package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sni/devices"
)

// putFileVerifyAfterWrite controls the post-transfer INFO round trip below.
// It exists so tests can measure the device's behavior with and without the
// extra command; production code should leave it enabled.
var putFileVerifyAfterWrite = true

type putFileRequest struct {
	path   string
	rom    []byte
	report devices.ProgressReportFunc
}

func (d *Device) putFile(ctx context.Context, path string, size uint32, r io.Reader, progress devices.ProgressReportFunc) (n uint32, err error) {
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
		d.lock.Lock()
		defer d.lock.Unlock()
	}

	// send command:
	err = sendSerialChunked(ctx, d.f, 512, sb)
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
		n, err = size, fmt.Errorf("putfile: response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if sb[4] != byte(OpRESPONSE) {
		n, err = size, fmt.Errorf("putfile: wrong opcode in response packet; got $%02x", sb[4])
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		// The device is now mid-transaction and must not be written to again.
		//
		// usbint_recv_block sets cmdDat=1 as soon as it sees a PUT opcode, before
		// usbint_handler_cmd has even attempted f_open. So a PUT that fails still
		// parks the firmware in HANDLE_LOCK waiting for the entire payload;
		// server_state only leaves HANDLE_LOCK once count >= server_info.size.
		// We are not sending that payload.
		//
		// Whatever is written next is therefore consumed as file data, and once a
		// full 512-byte block lands, f_write on the failed handle returns zero
		// bytes written while the loop waits for bytesRecv to reach block_size --
		// an infinite loop inside the USB interrupt handler. Confirmed on
		// hardware: one failed PUT plus one following command bricks the device
		// until it is physically power cycled.
		//
		// Reporting this as fatal makes autoCloseableDevice close and reopen the
		// device, and the firmware's usbint_check_connect() resets server_state
		// and cmdDat on disconnect. That is the only clean way out, and it only
		// works while nothing else has been written.
		n, err = 0, fmt.Errorf("putfile: %w", fxpakproError(ec))
		err = d.FatalError(err)
		return
	}

	if size == 0 {
		{
			tmp := make([]byte, 512)
			var m int
			m, err = d.f.Write(tmp)
			if debugLog != nil {
				debugLog.Printf("putFile: extra write: %#v, %#v\n", m, err)
			}
			_ = m
		}
		n = 0
		err = nil
		return
	}

	// send data:
	n, err = sendSerialProgress(ctx, d.f, 512, size, r, progress)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	// The fxpakpro sends exactly one response for a PUT, and it sends it before
	// the data phase, so nothing in the transfer itself tells us the device kept
	// up. A successful write() only means the host handed the bytes to the OS.
	// The firmware does its FAT work inside the USB interrupt handler, so if it
	// ever loses a packet mid-transfer it desyncs silently and we would report a
	// success here, leaving the failure to surface on some later unrelated
	// command. Round-trip an INFO to confirm the device is still responsive and
	// still framing the protocol correctly.
	//
	// d.lock is already held, so mark the context to keep info() from re-locking:
	if putFileVerifyAfterWrite {
		subctx := context.WithValue(ctx, lockedKey, &struct{}{})
		if _, _, _, verr := d.info(subctx); verr != nil {
			err = d.FatalError(fmt.Errorf("putfile: device did not respond after writing %d bytes to %#v: %w", n, path, verr))
			return
		}
	}

	return
}
