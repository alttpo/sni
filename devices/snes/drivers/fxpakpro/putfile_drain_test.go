package fxpakpro

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

// TestDevice_putFile_closeWithoutDrain tests whether closing the serial port
// straight after a PUT can wedge the device.
//
// putFile returns as soon as the last write() is buffered by the OS, and
// Device.Close() calls f.Close() without ever calling f.Drain(). If the close
// discards output the CDC driver had not yet transmitted, the firmware is left
// holding a partial 512-byte block in recv_buffer. usbint_check_connect()
// resets server_state, data_ready and cmdDat on disconnect but never resets
// recv_buffer_offset, so every command after the next connect is misframed by
// that leftover count and never assembles into a valid USBA block: the device
// enumerates, accepts bytes, and never answers.
//
// SNI_TEST_CLOSE_MODE selects what happens between the last data write and the
// close:
//
//	none  - close immediately (the current behavior)
//	drain - call Drain() first
//	info  - round-trip an INFO first (fix #2)
func TestDevice_putFile_closeWithoutDrain(t *testing.T) {
	mode := os.Getenv("SNI_TEST_CLOSE_MODE")
	if mode == "" {
		mode = "none"
	}
	t.Logf("close mode: %q", mode)

	const size = uint32(largeTestSize)
	const path = "unittest-snfm/sub/drain-test.bin"

	putFileVerifyAfterWrite = (mode == "info")
	defer func() { putFileVerifyAfterWrite = true }()

	// phase 1: upload, then close the port according to the mode.
	{
		d := openExactDevice(t)
		ctx := context.Background()

		if err := d.mkdir(ctx, "unittest-snfm"); err != nil {
			t.Logf("mkdir: %v (assuming exists)", err)
		}
		if err := d.mkdir(ctx, "unittest-snfm/sub"); err != nil {
			t.Logf("mkdir sub: %v (assuming exists)", err)
		}

		start := time.Now()
		n, err := d.putFile(ctx, path, size, bytes.NewReader(offsetPattern(size)), nil)
		if err != nil {
			d.Close()
			t.Fatalf("putFile: %v", err)
		}
		t.Logf("putFile: %d bytes in %v", n, time.Since(start))

		// SNI_TEST_GETFILE=1 adds a 4 MiB read-back before the close. The one
		// wedge observed in this session followed a test whose final device
		// operation was exactly this: a getFile, then Close(), then an idle
		// period, then a reopen that could not complete INFO.
		if os.Getenv("SNI_TEST_GETFILE") == "1" {
			var w bytes.Buffer
			w.Grow(int(size))
			gs := time.Now()
			received, gerr := d.getFile(ctx, path, &w, nil, nil)
			if gerr != nil {
				d.Close()
				t.Fatalf("getFile: %v", gerr)
			}
			t.Logf("getFile: %d bytes in %v", received, time.Since(gs))
		}

		if mode == "drain" {
			ds := time.Now()
			if err := d.f.Drain(); err != nil {
				t.Logf("Drain(): %v", err)
			}
			t.Logf("Drain() took %v", time.Since(ds))
		}

		// close immediately, exactly as SNI does when a device is released:
		if err := d.Close(); err != nil {
			t.Logf("Close(): %v", err)
		}
		t.Logf("closed port")
	}

	// phase 2: reopen and see whether the device still answers a command.
	// driver.openDevice runs Init(), which issues INFO -- the same command that
	// failed in the wedge we observed.
	uri, err := firstDeviceURI()
	if err != nil {
		t.Fatalf("detect: %v", err)
	}

	start := time.Now()
	dev, err := driver.openDevice(uri)
	if err != nil {
		t.Fatalf("REPRODUCED: reopen after close failed in %v: %v", time.Since(start), err)
	}
	d2 := dev.(*Device)
	defer d2.Close()
	t.Logf("reopened and INFO succeeded in %v", time.Since(start))

	ctx := context.Background()
	if _, err := d2.listFiles(ctx, "unittest-snfm/sub"); err != nil {
		t.Fatalf("REPRODUCED: ls after reopen failed: %v", err)
	}
	if err := d2.rm(ctx, path); err != nil {
		t.Logf("cleanup rm: %v", err)
	}
}
