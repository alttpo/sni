package fxpakpro

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestDevice_snfmSequence_autoCloseable runs the SNFM command sequence through
// devices.AutoCloseableDevice, the same wrapper services/grpcimpl uses, rather
// than against the driver directly. That layer is what turns a fatal error into
// a Close() plus DeleteDevice(), so the following request reopens the serial
// port -- and the fxpakpro firmware has no path to re-establish a USB session,
// which makes the reopen unrecoverable without a power cycle.
func TestDevice_snfmSequence_autoCloseable(t *testing.T) {
	verify := true
	if v := os.Getenv("SNI_TEST_PUTFILE_VERIFY"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			t.Fatalf("SNI_TEST_PUTFILE_VERIFY=%q: %v", v, err)
		}
		verify = b
	}
	old := putFileVerifyAfterWrite
	putFileVerifyAfterWrite = verify
	defer func() { putFileVerifyAfterWrite = old }()

	size := uint32(largeTestSize)
	if s := os.Getenv("SNI_TEST_PUTFILE_SIZE"); s != "" {
		v, err := strconv.ParseUint(s, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_PUTFILE_SIZE=%q: %v", s, err)
		}
		size = uint32(v)
	}

	iterations := 1
	if v := os.Getenv("SNI_TEST_ITERATIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SNI_TEST_ITERATIONS=%q: %v", v, err)
		}
		iterations = n
	}

	t.Logf("putFileVerifyAfterWrite=%v, size=%d, iterations=%d", verify, size, iterations)

	d := openAutoCloseableDevice(t)
	defer d.Close()

	ctx := context.Background()
	parents := []string{"unittest-snfm", "unittest-snfm/sub"}
	dir := parents[len(parents)-1]
	path := dir + "/snfm-test.bin"

	lsDir := dir
	if v := os.Getenv("SNI_TEST_LS_DIR"); v != "" {
		lsDir = v
	}

	expected := offsetPattern(size)

	for i := 0; i < iterations; i++ {
		// SNFM creates the folder and each parent first:
		for _, p := range parents {
			if err := d.MakeDirectory(ctx, p); err != nil {
				t.Logf("[%d] mkdir(%s): %v (assuming it exists)", i, p, err)
			}
		}

		start := time.Now()
		n, err := d.PutFile(ctx, path, size, bytes.NewReader(expected), nil)
		if err != nil {
			t.Fatalf("[%d] PutFile(%s, %d): %v", i, path, size, err)
		}
		if n != size {
			t.Fatalf("[%d] PutFile() sent %d bytes, want %d", i, n, size)
		}
		putDur := time.Since(start)

		lsStart := time.Now()
		files, err := d.ReadDirectory(ctx, lsDir)
		if err != nil {
			t.Fatalf("[%d] FAILURE: ReadDirectory(%s) after PutFile() failed after %v: %v",
				i, lsDir, time.Since(lsStart), err)
		}
		t.Logf("[%d] PutFile %v; ReadDirectory(%s) -> %d entries in %v",
			i, putDur, lsDir, len(files), time.Since(lsStart))

		if _, err := d.FetchFields(ctx, 0); err != nil {
			t.Fatalf("[%d] FAILURE: FetchFields() after ReadDirectory(): %v", i, err)
		}

		if err := d.RemoveFile(ctx, path); err != nil {
			t.Fatalf("[%d] RemoveFile(%s): %v", i, path, err)
		}
	}
}
