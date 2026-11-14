package fxpakpro

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

// Test_openPort_failsFastOnBadPort checks that opening a port which is not
// merely at the wrong speed gives up after one attempt instead of walking the
// whole baud table.
//
// A wedged fxpakpro makes each open attempt block for tens of seconds on
// Windows. Retrying all 14 rates cost roughly eight minutes of a caller sitting
// on a device that was never going to open, which is what this prevents.
func Test_openPort_failsFastOnBadPort(t *testing.T) {
	var buf bytes.Buffer
	old := log.Default().Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	d := &Driver{}

	start := time.Now()
	f, err := d.openPort("/dev/nonexistent-fxpakpro-port-for-test", baudRates[0])
	elapsed := time.Since(start)

	if err == nil {
		if f != nil {
			_ = f.Close()
		}
		t.Fatalf("openPort() on a nonexistent port returned no error")
	}

	attempts := strings.Count(buf.String(), "open(name=\"/dev/nonexistent-fxpakpro-port-for-test\", baud=")
	if attempts != 1 {
		t.Errorf("openPort() tried %d baud rates; want 1 (log:\n%s)", attempts, buf.String())
	}
	if elapsed > 10*time.Second {
		t.Errorf("openPort() took %v to fail", elapsed)
	}
	t.Logf("failed after %d attempt(s) in %v: %v", attempts, elapsed, err)
}
