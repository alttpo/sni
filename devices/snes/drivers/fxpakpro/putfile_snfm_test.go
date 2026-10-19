package fxpakpro

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestDevice_putFile_thenLS reproduces the command sequence the SNFM file
// transfer tool issues: MKDIR for the target folder and each of its parents,
// then PutFile, then LS of the folder it just wrote into.
//
// LS is the interesting part. usbint_handler_cmd sets server_state to
// USBINT_SERVER_STATE_HANDLE_DAT for LS (as it does for GET/VGET), and
// usbint_server_busy() counts HANDLE_DAT as busy -- unlike PUT's HANDLE_LOCK,
// which it deliberately excludes. While the server is busy, CDC_BulkOut()
// returns without calling USB_ReadEP(), so the packet the USB hardware already
// ACKed is discarded and its endpoint buffer is never released with
// CMD_CLR_BUF. The endpoint interrupt was already cleared by the ISR before the
// callback ran, so nothing brings that buffer back.
//
// SNI_TEST_PUTFILE_VERIFY=0 disables the post-PUT INFO round trip in putFile()
// so the two behaviors can be compared.
func TestDevice_putFile_thenLS(t *testing.T) {
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

	t.Logf("putFileVerifyAfterWrite=%v, size=%d bytes (%.1f MiB)",
		verify, size, float64(size)/1024.0/1024.0)

	d := openExactDevice(t)
	t.Cleanup(func() { d.Close() })
	ctx := context.Background()

	// SNFM creates the target folder and each parent in turn:
	parents := []string{"unittest-snfm", "unittest-snfm/sub"}
	dir := parents[len(parents)-1]
	for _, p := range parents {
		if err := d.mkdir(ctx, p); err != nil {
			// error code 1 just means it already exists:
			if _, lserr := d.listFiles(ctx, p); lserr != nil {
				t.Fatalf("mkdir(%s): %v (and it does not already exist: %v)", p, err, lserr)
			}
			t.Logf("mkdir(%s): already exists", p)
		}
	}

	// LS target: default the folder we uploaded into, as SNFM does. A folder
	// with many entries makes the LS data phase -- and therefore the window in
	// which usbint_server_busy() reports busy -- much longer.
	lsDir := dir
	if v := os.Getenv("SNI_TEST_LS_DIR"); v != "" {
		lsDir = v
	}

	iterations := 1
	if v := os.Getenv("SNI_TEST_ITERATIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SNI_TEST_ITERATIONS=%q: %v", v, err)
		}
		iterations = n
	}
	t.Logf("ls target %q, %d iteration(s)", lsDir, iterations)

	path := dir + "/snfm-test.bin"
	t.Cleanup(func() {
		if err := d.rm(context.Background(), path); err != nil {
			t.Logf("cleanup: rm(%s): %v", path, err)
		}
	})

	expected := offsetPattern(size)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		n, err := d.putFile(ctx, path, size, bytes.NewReader(expected), nil)
		if err != nil {
			t.Fatalf("[%d] putFile(%s, %d): %v", i, path, size, err)
		}
		if n != size {
			t.Fatalf("[%d] putFile() sent %d bytes, want %d", i, n, size)
		}
		putDur := time.Since(start)

		// LS immediately, exactly as SNFM does. This is the command reported to
		// fail, so give it its own error handling.
		lsStart := time.Now()
		files, err := d.listFiles(ctx, lsDir)
		if err != nil {
			t.Fatalf("[%d] FAILURE: ls(%s) immediately after putFile() failed after %v: %v",
				i, lsDir, time.Since(lsStart), err)
		}
		t.Logf("[%d] putFile %v; ls(%s) -> %d entries in %v",
			i, putDur, lsDir, len(files), time.Since(lsStart))

		if lsDir == dir {
			var found bool
			for _, f := range files {
				if f.Name == "snfm-test.bin" {
					found = true
				}
			}
			if !found {
				t.Fatalf("[%d] ls(%s) did not list the file just uploaded", i, dir)
			}
		}

		// a second command afterwards catches a device left in a bad state:
		if _, _, _, err := d.info(ctx); err != nil {
			t.Fatalf("[%d] FAILURE: info() after ls() failed: %v", i, err)
		}

		if err := d.rm(ctx, path); err != nil {
			t.Fatalf("[%d] rm(%s) between iterations: %v", i, path, err)
		}
	}
}
