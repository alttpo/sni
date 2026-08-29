package fxpakpro

import (
	"context"
	"testing"
	"time"

	"sni/cmd/sni/config"
)

// contextExpired returns a context whose deadline has already passed.
func contextExpired() (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
}

func Test_loadTimeoutConfig(t *testing.T) {
	oldRead, oldWrite, oldHonor := noDataTimeout, writeTimeout, honorCallerDeadline
	defer func() {
		noDataTimeout, writeTimeout, honorCallerDeadline = oldRead, oldWrite, oldHonor
	}()

	config.Config.Set("fxpakpro_read_timeout", "42s")
	config.Config.Set("fxpakpro_write_timeout", "7s")
	config.Config.Set("fxpakpro_honor_caller_deadline", false)

	loadTimeoutConfig()

	if noDataTimeout != 42*time.Second {
		t.Errorf("noDataTimeout = %v; want 42s", noDataTimeout)
	}
	if writeTimeout != 7*time.Second {
		t.Errorf("writeTimeout = %v; want 7s", writeTimeout)
	}
	if honorCallerDeadline {
		t.Errorf("honorCallerDeadline = true; want false")
	}

	// a nonsense value must not disable the timeout entirely, which would
	// reintroduce the unbounded wait these settings exist to prevent:
	config.Config.Set("fxpakpro_read_timeout", "0s")
	config.Config.Set("fxpakpro_write_timeout", "-5s")
	loadTimeoutConfig()

	if noDataTimeout != 42*time.Second {
		t.Errorf("noDataTimeout = %v after a 0s setting; want the previous value kept", noDataTimeout)
	}
	if writeTimeout != 7*time.Second {
		t.Errorf("writeTimeout = %v after a negative setting; want the previous value kept", writeTimeout)
	}
}

// Test_honorCallerDeadline_off checks that with the policy disabled, a caller's
// expired deadline no longer aborts a write; only the driver's own budget does.
func Test_honorCallerDeadline_off(t *testing.T) {
	oldHonor, oldWrite := honorCallerDeadline, writeTimeout
	honorCallerDeadline = false
	writeTimeout = 150 * time.Millisecond
	defer func() { honorCallerDeadline, writeTimeout = oldHonor, oldWrite }()

	p := &blockingPort{release: make(chan struct{}), started: make(chan struct{})}
	defer close(p.release)

	// already-expired context: with the policy on this returns immediately.
	ctx, cancel := contextExpired()
	defer cancel()

	start := time.Now()
	_, err := writeExact(ctx, p, 512, make([]byte, 512))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("writeExact() returned nil error; want the write-timeout error")
	}
	if elapsed < 100*time.Millisecond {
		t.Errorf("writeExact() returned after %v; with the caller deadline ignored "+
			"it should have waited for the %v budget", elapsed, writeTimeout)
	}
}
