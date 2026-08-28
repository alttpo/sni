package fxpakpro

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.bug.st/serial"
)

// failingPort is a serial.Port that fails the failOnCall'th Write and succeeds
// on every other one. A transient failure is what actually exposes the bug: a
// permanently broken port leaves err set on the final loop iteration, so it
// gets returned by accident. When a later write succeeds, the old code's
// readExactGeneric reassigned err and the failure vanished.
type failingPort struct {
	stubPort
	calls      int
	written    int
	failOnCall int
	err        error
}

func (p *failingPort) Write(b []byte) (int, error) {
	p.calls++
	if p.calls == p.failOnCall {
		return 0, p.err
	}
	p.written += len(b)
	return len(b), nil
}

// stubPort supplies the parts of serial.Port the tests do not care about.
type stubPort struct{}

func (stubPort) Read(b []byte) (int, error) { return 0, nil }
func (stubPort) SetMode(*serial.Mode) error { return nil }
func (stubPort) Drain() error               { return nil }
func (stubPort) ResetInputBuffer() error    { return nil }
func (stubPort) ResetOutputBuffer() error   { return nil }
func (stubPort) SetDTR(bool) error          { return nil }
func (stubPort) SetRTS(bool) error          { return nil }
func (stubPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}
func (stubPort) SetReadTimeout(time.Duration) error { return nil }
func (stubPort) Close() error                       { return nil }
func (stubPort) Break(time.Duration) error          { return nil }

// Test_sendSerialProgress_writeError checks that a write failure partway
// through a transfer is reported rather than swallowed. The loop previously
// discarded writeExact's error and kept going, so the next iteration's
// readExactGeneric reassigned err and the caller saw a success for a transfer
// that had actually stopped writing.
func Test_sendSerialProgress_writeError(t *testing.T) {
	const size = 8192
	// fail the 5th chunk, so 4 chunks (2048 bytes) go out first and the
	// remaining chunks would have succeeded:
	const failOnCall = 5
	const wantSent = (failOnCall - 1) * 512

	wantErr := errors.New("simulated write failure")
	p := &failingPort{failOnCall: failOnCall, err: wantErr}

	sent, err := sendSerialProgress(context.Background(), p, 512, size, bytes.NewReader(make([]byte, size)), nil)
	if err == nil {
		t.Fatalf("sendSerialProgress() returned nil error after a failed write; want %v (sent=%d of %d)",
			wantErr, sent, size)
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("sendSerialProgress() error = %v; want it to wrap %v", err, wantErr)
	}
	if sent != wantSent {
		t.Errorf("sendSerialProgress() sent = %d; want %d (should stop at the failed write)", sent, wantSent)
	}
}

// Test_sendSerialProgress_success covers the ordinary path, including a size
// that is not a multiple of the chunk size so the remainder branch runs.
func Test_sendSerialProgress_success(t *testing.T) {
	for _, size := range []uint32{512, 8192, 513, 1023} {
		p := &failingPort{failOnCall: -1}
		sent, err := sendSerialProgress(context.Background(), p, 512, size, bytes.NewReader(make([]byte, size)), nil)
		if err != nil {
			t.Errorf("size %d: sendSerialProgress() error = %v", size, err)
			continue
		}
		// the protocol pads the final short chunk out to a full 512 bytes:
		want := ((size + 511) / 512) * 512
		if sent != want {
			t.Errorf("size %d: sent = %d; want %d", size, sent, want)
		}
	}
}

// blockingPort models a device that has stopped draining its USB endpoint: the
// write never completes until the port is closed.
type blockingPort struct {
	stubPort
	release chan struct{}
	started chan struct{}
	once    sync.Once
}

func (p *blockingPort) Write(b []byte) (int, error) {
	p.once.Do(func() { close(p.started) })
	<-p.release
	return 0, errors.New("port closed")
}

// Test_writeWithTimeout_deviceNotDraining checks that a write to a device that
// never accepts data returns instead of hanging forever. go.bug.st/serial sets
// WriteTotalTimeoutConstant to 0 on Windows, which Win32 treats as "wait
// forever", and the Port interface offers no SetWriteTimeout, so the bound has
// to come from us. Without it the caller hangs while holding d.lock, blocking
// every other request for that device.
func Test_writeWithTimeout_deviceNotDraining(t *testing.T) {
	old := writeTimeout
	writeTimeout = 150 * time.Millisecond
	defer func() { writeTimeout = old }()

	p := &blockingPort{release: make(chan struct{}), started: make(chan struct{})}
	// release the abandoned goroutine at the end, as closing the port would:
	defer close(p.release)

	start := time.Now()
	n, err := writeExact(context.Background(), p, 512, make([]byte, 512))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("writeExact() returned nil error after %v; want a timeout", elapsed)
	}
	if n != 0 {
		t.Errorf("writeExact() = %d bytes; want 0 when the write never completed", n)
	}
	if elapsed > 5*time.Second {
		t.Errorf("writeExact() took %v; it should give up near the %v budget", elapsed, writeTimeout)
	}

	select {
	case <-p.started:
	default:
		t.Errorf("Write() was never called")
	}
}

// Test_writeWithTimeout_contextCancelled checks that a caller's context bounds
// the write too, so a cancelled request does not sit on the device lock.
func Test_writeWithTimeout_contextCancelled(t *testing.T) {
	old := writeTimeout
	writeTimeout = time.Minute // ensure the context is what ends the wait
	defer func() { writeTimeout = old }()

	p := &blockingPort{release: make(chan struct{}), started: make(chan struct{})}
	defer close(p.release)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := writeExact(ctx, p, 512, make([]byte, 512))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("writeExact() returned nil error after %v; want context deadline exceeded", elapsed)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("writeExact() error = %v; want it to wrap context.DeadlineExceeded", err)
	}
	if elapsed > 5*time.Second {
		t.Errorf("writeExact() took %v; it should stop at the context deadline", elapsed)
	}
}
