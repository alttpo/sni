package fxpakpro

import (
	"context"
	"testing"
	"time"
)

func expiredContext() (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
}

// TestDevice_deadlinePolicy exercises fxpakpro_honor_caller_deadline against
// the real device. With the policy on, a caller whose deadline has already
// passed must not get to talk to the device; with it off, the same call must
// run to completion under the driver's own timeouts instead.
func TestDevice_deadlinePolicy(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	restore := honorCallerDeadline
	defer func() { honorCallerDeadline = restore }()

	// baseline: the device is healthy before we start
	if _, _, rom, err := d.info(context.Background()); err != nil {
		t.Fatalf("baseline info(): %v", err)
	} else {
		t.Logf("baseline ok, rom=%q", rom)
	}

	// The policy-off case runs first: honoring an expired deadline abandons a
	// write, which closes the port, so that case has to come last.
	t.Run("ignored", func(t *testing.T) {
		honorCallerDeadline = false

		ctx, cancel := expiredContext()
		defer cancel()

		start := time.Now()
		version, _, rom, err := d.info(ctx)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("info() with the deadline policy off failed after %v: %v", elapsed, err)
		}
		if version == "" {
			t.Fatalf("info() returned an empty version; the stream is out of step")
		}
		t.Logf("completed in %v despite the expired deadline: version=%q rom=%q",
			elapsed, version, rom)
	})

	t.Run("honored", func(t *testing.T) {
		honorCallerDeadline = true

		ctx, cancel := expiredContext()
		defer cancel()

		start := time.Now()
		_, _, _, err := d.info(ctx)
		elapsed := time.Since(start)
		if err == nil {
			t.Fatalf("info() with an expired deadline succeeded after %v; want it aborted", elapsed)
		}
		t.Logf("aborted after %v: %v", elapsed, err)
		if elapsed > 5*time.Second {
			t.Errorf("abort took %v; an already-expired deadline should return promptly", elapsed)
		}

		// Abandoning the write must have closed the port, so the orphaned
		// goroutine cannot interleave with a later command. The next call has to
		// fail immediately rather than sitting through the write timeout.
		nstart := time.Now()
		_, _, _, nerr := d.info(context.Background())
		nelapsed := time.Since(nstart)
		if nerr == nil {
			t.Errorf("a command succeeded after an abandoned write; the port should be closed")
		} else {
			t.Logf("next command failed in %v as expected: %v", nelapsed, nerr)
		}
		if nelapsed > 5*time.Second {
			t.Errorf("next command took %v; a closed port should fail immediately, "+
				"not sit through the %v write timeout", nelapsed, writeTimeout)
		}
	})
}
