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

// TestDevice_repeatLargeTransfer repeatedly PUTs and GETs a single large file,
// counting how many round trips the device survives.
//
// Every wedge observed so far involved a large transfer: getFile at 188416 and
// 198201 bytes, and putFile at 131071 bytes. This narrows the random stress mix
// down to just that, to find out whether large transfers alone are the trigger
// and how many it takes -- a far tighter reproduction to hand to the firmware
// than "run 6000 random operations".
//
// SNI_TEST_XFER_SIZE sets the payload size, so the threshold can be bisected.
func TestDevice_repeatLargeTransfer(t *testing.T) {
	size := uint32(128 * 1024)
	if v := os.Getenv("SNI_TEST_XFER_SIZE"); v != "" {
		n, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_XFER_SIZE=%q: %v", v, err)
		}
		size = uint32(n)
	}
	iterations := 200
	if v := os.Getenv("SNI_TEST_ITERATIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SNI_TEST_ITERATIONS=%q: %v", v, err)
		}
		iterations = n
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	const dir = "unittest-large-xfer"
	if err := d.mkdir(ctx, dir); err != nil {
		if _, lserr := d.listFiles(ctx, dir); lserr != nil {
			t.Fatalf("mkdir(%s): %v (and it does not exist: %v)", dir, err, lserr)
		}
	}

	t.Logf("payload %d bytes (%.1f KiB), up to %d round trips",
		size, float64(size)/1024.0, iterations)

	payload := filePattern(0x5a5a5a5a, size)
	path := dir + "/xfer.bin"
	start := time.Now()

	for i := 1; i <= iterations; i++ {
		putStart := time.Now()
		n, err := d.putFile(ctx, path, size, bytes.NewReader(payload), nil)
		if err != nil {
			t.Fatalf("WEDGED on round trip %d after %v: putFile: %v", i, time.Since(start), err)
		}
		if n != size {
			t.Fatalf("round trip %d: putFile sent %d bytes, want %d", i, n, size)
		}
		putDur := time.Since(putStart)

		getStart := time.Now()
		var w bytes.Buffer
		w.Grow(int(size))
		received, err := d.getFile(ctx, path, &w, nil, nil)
		if err != nil {
			t.Fatalf("WEDGED on round trip %d after %v: getFile: %v", i, time.Since(start), err)
		}
		if received != size {
			t.Fatalf("round trip %d: getFile received %d bytes, want %d", i, received, size)
		}
		if !bytes.Equal(w.Bytes(), payload) {
			t.Fatalf("round trip %d: %s", i,
				describeStressMismatch(w.Bytes(), payload, 0x5a5a5a5a))
		}
		getDur := time.Since(getStart)

		if err := d.rm(ctx, path); err != nil {
			t.Fatalf("WEDGED on round trip %d after %v: rm: %v", i, time.Since(start), err)
		}

		if i%10 == 0 || i <= 5 {
			t.Logf("round trip %3d ok (put %v, get %v, total elapsed %v)",
				i, putDur.Round(time.Millisecond), getDur.Round(time.Millisecond),
				time.Since(start).Round(time.Millisecond))
		}
	}

	t.Logf("survived %d round trips of %s in %v",
		iterations, fmt.Sprintf("%d bytes", size), time.Since(start))
}
