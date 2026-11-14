package fxpakpro

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestDevice_fillUntilFull writes files until the card runs out of space, to
// find out what the device does at the boundary.
//
// There are two distinct disk-full cases and they behave very differently:
//
//   - f_open fails up front, the device answers with an error code, and SNI can
//     bail out cleanly. putFile now treats that as fatal so the connection is
//     torn down inside the one-command window the firmware allows.
//   - f_open succeeds and f_write fails partway through the data phase. There is
//     no error code to see: usbint_recv_block loops
//     `while (bytesRecv != block_size && count < size)` and adds bytesWritten,
//     which is 0 once the card is full, so it never advances -- an infinite loop
//     inside the USB interrupt handler. The device stops draining its OUT
//     endpoint and the host blocks mid-transfer.
//
// The second is unrecoverable and is the better match for an end user uploading
// a 4 MiB ROM to a card packed with MSU tracks.
func TestDevice_fillUntilFull(t *testing.T) {
	if os.Getenv("SNI_TEST_FILL") == "" {
		t.Skip("set SNI_TEST_FILL=1; this fills the SD card and can wedge the device")
	}

	size := uint32(4 * 1024 * 1024)
	if v := os.Getenv("SNI_TEST_XFER_SIZE"); v != "" {
		n, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_XFER_SIZE=%q: %v", v, err)
		}
		size = uint32(n)
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	const dir = "unittest-fill"
	if err := d.mkdir(ctx, dir); err != nil {
		if _, lserr := d.listFiles(ctx, dir); lserr != nil {
			t.Fatalf("mkdir(%s): %v (and it does not exist: %v)", dir, err, lserr)
		}
	}

	payload := filePattern(0xa5a5a5a5, size)
	t.Logf("filling with %d byte files", size)

	start := time.Now()
	for i := 1; ; i++ {
		path := fmt.Sprintf("%s/fill%04d.bin", dir, i)

		putStart := time.Now()
		n, err := d.putFile(ctx, path, size, bytes.NewReader(payload), nil)
		putDur := time.Since(putStart)

		if err == nil {
			if i%10 == 0 {
				t.Logf("file %d written (%v each, %v total)", i,
					putDur.Round(time.Millisecond), time.Since(start).Round(time.Second))
			}
			continue
		}

		// This is the interesting part: how did it fail?
		t.Logf("file %d FAILED after %v having sent %d of %d bytes", i, putDur, n, size)
		t.Logf("error: %v", err)

		switch {
		case n == 0:
			t.Logf("=> clean rejection before the data phase (device answered with "+
				"an error code); %d files written in %v", i-1, time.Since(start))
		case n < size:
			t.Logf("=> WEDGED MID-TRANSFER: the device accepted the command, then "+
				"stopped draining after %d of %d bytes. This is the f_write "+
				"disk-full case: no error code to see, and the firmware is "+
				"spinning in its USB interrupt handler.", n, size)
		}

		// is it still alive?
		probe := time.Now()
		if _, _, _, ierr := d.info(ctx); ierr != nil {
			t.Fatalf("device unresponsive %v after the failure: %v", time.Since(probe), ierr)
		}
		t.Logf("device still responds after the failure (%v)", time.Since(probe))
		return
	}
}
