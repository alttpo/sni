package fxpakpro

import (
	"context"
	"sni/devices"
	"testing"
	"time"
)

func expiredContext() (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
}

// TestDevice_deadlinePolicy exercises fxpakpro_honor_caller_deadline against
// the real device. Either way, a command that has started must run to the end
// of its protocol exchange so the device stays in step; the policy only
// decides whether the caller waits for that.
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

	t.Run("honored/expired", func(t *testing.T) {
		honorCallerDeadline = true

		ctx, cancel := expiredContext()
		defer cancel()

		start := time.Now()
		_, _, _, err := d.info(ctx)
		elapsed := time.Since(start)
		if err == nil {
			t.Fatalf("info() with an expired deadline succeeded after %v; want it refused", elapsed)
		}
		if devices.IsFatal(err) {
			t.Errorf("info() error is fatal: %v; a caller's deadline must not close the device", err)
		}
		if elapsed > time.Second {
			t.Errorf("refusal took %v; an already-expired deadline should return promptly", elapsed)
		}

		// nothing was sent, so the device is untouched:
		if _, _, _, err := d.info(context.Background()); err != nil {
			t.Fatalf("info() after an expired caller: %v", err)
		}
	})

	// Deadlines short enough to land partway through a command. Each caller
	// must get its error, and the device must still answer correctly
	// afterwards: the abandoned commands were drained, not cut off.
	t.Run("honored/midCommand", func(t *testing.T) {
		honorCallerDeadline = true

		abandoned := 0
		for i := 0; i < 50; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(i%10)*100*time.Microsecond+50*time.Microsecond)
			_, _, _, err := d.info(ctx)
			cancel()
			if err != nil {
				if devices.IsFatal(err) {
					t.Fatalf("[%d] info() error is fatal: %v", i, err)
				}
				abandoned++
			}
		}
		t.Logf("%d of 50 commands were abandoned by their caller", abandoned)

		for i := 0; i < 5; i++ {
			version, device, _, err := d.info(context.Background())
			if err != nil {
				t.Fatalf("[%d] info() after abandoned commands: %v", i, err)
			}
			if version == "" || device == "" {
				t.Fatalf("[%d] info() returned an empty field (version=%q device=%q); the stream is out of step",
					i, version, device)
			}
		}
	})
}
