package fxpakpro

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestDevice_usbstatProbe clears the firmware's counters file, runs a chosen
// workload, then reads the counters back, so drops can be attributed to
// commands rather than to bulk data.
//
// SNI_TEST_PROBE selects the workload:
//
//	info  - N INFO commands: pure command traffic, no data phase
//	put   - one PutFile of SNI_TEST_XFER_SIZE: one command, thousands of
//	        512-byte data blocks
//
// If drops scale with the number of commands they happen at the command
// boundary; if they scale with data volume they happen during the transfer.
// Note the bookkeeping calls (rm, ls, get) are themselves commands and are
// included in the totals.
func TestDevice_usbstatProbe(t *testing.T) {
	probe := os.Getenv("SNI_TEST_PROBE")
	if probe == "" {
		t.Skip("set SNI_TEST_PROBE to info or put")
	}
	n := 10
	if v := os.Getenv("SNI_TEST_ITERATIONS"); v != "" {
		var err error
		if n, err = strconv.Atoi(v); err != nil {
			t.Fatalf("SNI_TEST_ITERATIONS=%q: %v", v, err)
		}
	}
	size := uint32(4 * 1024 * 1024)
	if v := os.Getenv("SNI_TEST_XFER_SIZE"); v != "" {
		u, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_XFER_SIZE=%q: %v", v, err)
		}
		size = uint32(u)
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	if err := d.rm(ctx, "sd2snes/usbstat.txt"); err != nil {
		t.Logf("clear counters: %v (file may not exist yet)", err)
	}

	switch probe {
	case "info":
		t.Logf("workload: %d INFO commands (no data phase)", n)
		for i := 0; i < n; i++ {
			if _, _, _, err := d.info(ctx); err != nil {
				t.Fatalf("info %d: %v", i, err)
			}
		}
	case "put":
		t.Logf("workload: one PutFile of %d bytes (%d data blocks)", size, size/512)
		payload := filePattern(0x11223344, size)
		if err := d.mkdir(ctx, "unittest"); err != nil {
			t.Logf("mkdir: %v (assuming exists)", err)
		}
		if _, err := d.putFile(ctx, "unittest/usbstat-probe.bin", size,
			bytes.NewReader(payload), nil); err != nil {
			t.Fatalf("putFile: %v", err)
		}
	default:
		t.Fatalf("SNI_TEST_PROBE=%q: want info or put", probe)
	}

	// give the firmware an idle pass to flush, then read the counters
	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		_, _, _, _ = d.info(ctx)
	}

	var w bytes.Buffer
	if _, err := d.getFile(ctx, "sd2snes/usbstat.txt", &w, nil, nil); err != nil {
		t.Logf("no counters file: %v (no drop was recorded)", err)
		return
	}
	for _, line := range strings.Split(strings.TrimRight(w.String(), "\x00\n"), "\n") {
		if line != "" {
			t.Logf("  %s", line)
		}
	}
}
