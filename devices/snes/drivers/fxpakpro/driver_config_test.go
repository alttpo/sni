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
