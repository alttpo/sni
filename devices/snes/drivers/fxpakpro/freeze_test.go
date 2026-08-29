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

// TestDevice_freezeHunt looks for the reported symptom: SNI hanging during a
// PutFile because the device stopped draining its USB OUT endpoint.
//
// That is what the end user saw. gRPC PutFile is unary and blocking with no
// progress reporting, so a stalled write presents simply as SNI freezing. It
// was captured once as a goroutine sitting 29 minutes inside WriteFile, but
// every reproduction so far has been on a card with no free space, which is not
// the reporter's situation. This hunts for it with space available.
//
// The write timeout now turns such a stall into an error rather than a hang, so
// what this looks for is that error, plus any chunk that took implausibly long
// without failing outright -- a near miss is as interesting as a stall.
func TestDevice_freezeHunt(t *testing.T) {
	size := uint32(4 * 1024 * 1024)
	if v := os.Getenv("SNI_TEST_XFER_SIZE"); v != "" {
		n, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_XFER_SIZE=%q: %v", v, err)
		}
		size = uint32(n)
	}
	iterations := 40
	if v := os.Getenv("SNI_TEST_ITERATIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SNI_TEST_ITERATIONS=%q: %v", v, err)
		}
		iterations = n
	}
	// report any chunk slower than this; the worst seen on a healthy card is
	// around 130ms, so anything far above it is heading toward a stall
	stallWarn := 2 * time.Second
	if v := os.Getenv("SNI_TEST_STALL_WARN"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			t.Fatalf("SNI_TEST_STALL_WARN=%q: %v", v, err)
		}
		stallWarn = d
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	const dir = "unittest-freeze"
	if err := d.mkdir(ctx, dir); err != nil {
		if _, lserr := d.listFiles(ctx, dir); lserr != nil {
			t.Fatalf("mkdir(%s): %v (and it does not exist: %v)", dir, err, lserr)
		}
	}
	path := dir + "/freeze.bin"
	t.Cleanup(func() {
		if err := d.rm(context.Background(), path); err != nil {
			t.Logf("cleanup: rm(%s): %v", path, err)
		}
	})

	payload := filePattern(0x3c3c3c3c, size)
	t.Logf("%d iterations of %d bytes, warning on any chunk over %v",
		iterations, size, stallWarn)

	var worst time.Duration
	var worstAt uint32
	start := time.Now()

	for i := 1; i <= iterations; i++ {
		// progress callback lets us time the gap between chunks without
		// replacing the production write path, which is the point: this must
		// exercise sendSerialProgress exactly as PutFile does.
		last := time.Now()
		var slow []string
		progress := func(sent, total uint32) {
			if gap := time.Since(last); gap > stallWarn {
				slow = append(slow, fmt.Sprintf("%v at offset %d", gap.Round(time.Millisecond), sent))
			} else if gap > worst {
				worst, worstAt = gap, sent
			}
			last = time.Now()
		}

		iterStart := time.Now()
		n, err := d.putFile(ctx, path, size, bytes.NewReader(payload), progress)
		iterDur := time.Since(iterStart)

		if err != nil {
			t.Fatalf("FROZE on iteration %d after %v (%v total): sent %d of %d: %v",
				i, iterDur, time.Since(start), n, size, err)
		}
		for _, s := range slow {
			t.Errorf("iteration %d: STALL %s", i, s)
		}
		if i%5 == 0 {
			t.Logf("iteration %2d ok (%v, worst inter-chunk gap so far %v at offset %d)",
				i, iterDur.Round(time.Millisecond), worst.Round(time.Millisecond), worstAt)
		}
	}

	t.Logf("survived %d iterations in %v; worst inter-chunk gap %v at offset %d",
		iterations, time.Since(start).Round(time.Second), worst.Round(time.Millisecond), worstAt)
}
