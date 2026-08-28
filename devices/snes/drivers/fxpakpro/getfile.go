package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sni/devices"
)

func (d *Device) getFile(ctx context.Context, path string, w io.Writer, sizeReceived devices.SizeReceivedFunc, progress devices.ProgressReportFunc) (received uint32, err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpGET)
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
		received, err = 0, fmt.Errorf("getFile: response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if sb[4] != byte(OpRESPONSE) {
		received, err = 0, fmt.Errorf("getFile: wrong opcode in response packet; got $%02x", sb[4])
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		// Like PUT, the firmware commits to a data phase before it knows whether
		// the command succeeded: usbint_handler_cmd sets HANDLE_DAT for GET
		// regardless of whether f_stat and f_open worked. Unlike PUT, this one
		// cannot be recovered from the host side.
		//
		// server_info.size is taken from fi.fsize, and a failed f_stat never
		// writes fi -- it is a stale global shared with f_readdir. If it holds a
		// non-zero value, usbint_handler_dat spins forever: f_read on the failed
		// handle returns zero bytes while the loop waits for bytesSent to reach
		// block_size. The main loop never gets back to usbint_check_connect(), so
		// disconnecting does not help, and usbint_server_busy() reports HANDLE_DAT
		// as busy, so CDC_BulkOut stops draining the OUT endpoint and writes from
		// the host block too. Confirmed on hardware: this needs a power cycle.
		//
		// Report it as fatal so SNI closes the device and stops issuing commands
		// into a pak that cannot answer, and so the failure is attributed here
		// rather than to whatever unrelated request came next.
		received, err = 0, fmt.Errorf(
			"getFile: %w (a GET for a file the device cannot open may hang its "+
				"firmware until power cycled)", fxpakproError(ec))
		err = d.FatalError(err)
		return
	}

	// read the size of the file:
	size := binary.BigEndian.Uint32(sb[252:256])
	if sizeReceived != nil {
		sizeReceived(size)
	}

	if size == 0 {
		// FXPAKPRO BUG: GET for 0-byte file causes all subsequent reads to fail!
		//{
		//	tmp := make([]byte, 512)
		//	var m int
		//	m, err = d.f.Read(tmp)
		//	log.Printf("getFile: extra read: %#v, %#v\n", m, err)
		//}
		// no extra data to expect:
		received = 0
		err = nil
		return
	}

	// read all remaining bytes in chunks of 512 bytes:
	chunk := make([]byte, 512)
	chunkCount := size / 512

	received = 0
	if progress != nil {
		progress(received, size)
	}
	for i := uint32(0); i < chunkCount; i++ {
		_, err = readExact(ctx, d.f, 512, chunk)
		if err != nil {
			received = 0
			err = d.FatalError(err)
			return
		}

		var n int
		n, err = w.Write(chunk)
		if err != nil {
			err = d.NonFatalError(err)
			return
		}
		if n != 512 {
			err = d.NonFatalError(fmt.Errorf("fxpakpro: getFile: wrote only %d bytes out of %d byte chunk to io.Writer", n, 512))
			return
		}
		received += 512

		if progress != nil {
			progress(received, size)
		}
	}

	remainder := int(size & 511)
	if remainder != 0 {
		_, err = readExact(ctx, d.f, 512, chunk)
		if err != nil {
			received = 0
			err = d.FatalError(err)
			return
		}

		var n int
		n, err = w.Write(chunk[:remainder])
		if err != nil {
			err = d.NonFatalError(err)
			return
		}
		if n != remainder {
			err = d.NonFatalError(fmt.Errorf("fxpakpro: getFile: wrote only %d bytes out of %d byte chunk to io.Writer", n, remainder))
			return
		}
		received += uint32(remainder)

		if progress != nil {
			progress(received, size)
		}
	}

	return
}
