package fxpakpro

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"sni/devices"

	"google.golang.org/grpc/codes"
)

// fakePak is a serial.Port that answers INFO the way the firmware does: one
// 512-byte response per 512-byte command, strictly in order. Each response
// carries a sequence number in its ROM name field, so a test can tell whether
// a command received its own response or a stale one left behind by an
// earlier command.
//
// A response can be held back by setting gate, which models a device that is
// slow to answer (FatFs work inside the USB interrupt handler, say) long
// enough for a caller's deadline to pass mid-command.
type fakePak struct {
	stubPort

	mu          sync.Mutex
	in          []byte // command bytes not yet parsed
	out         []byte // response bytes not yet read by the host
	pending     int    // responses the device owes the host and has not finished sending
	commands    int
	violations  []string
	readTimeout time.Duration

	// gate, when non-nil, holds back the response to the next command until
	// it is closed. It applies to one command only.
	gate chan struct{}
}

func (p *fakePak) SetReadTimeout(t time.Duration) error {
	p.mu.Lock()
	p.readTimeout = t
	p.mu.Unlock()
	return nil
}

func (p *fakePak) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.in = append(p.in, b...)
	for len(p.in) >= 512 {
		cmd := p.in[:512]
		p.in = p.in[512:]

		// The protocol is strictly synchronous: the host must not start a new
		// command while the device still owes it data for the previous one.
		if p.pending > 0 || len(p.out) > 0 {
			p.violations = append(p.violations, fmt.Sprintf(
				"command %d sent while %d response(s) and %d byte(s) from the previous command were outstanding",
				p.commands, p.pending, len(p.out)))
		}
		if string(cmd[:4]) != "USBA" || opcode(cmd[4]) != OpINFO {
			p.violations = append(p.violations, fmt.Sprintf(
				"command %d is not a well-framed INFO: % x", p.commands, cmd[:8]))
			continue
		}

		rsp := make([]byte, 512)
		copy(rsp, "USBA")
		rsp[4] = byte(OpRESPONSE)
		copy(rsp[16:], fmt.Sprintf("seq%d", p.commands))
		copy(rsp[260:], "fake-version")
		copy(rsp[260+64:], "fake-device")
		p.commands++
		p.pending++

		if gate := p.gate; gate != nil {
			p.gate = nil
			go func() {
				<-gate
				p.mu.Lock()
				p.out = append(p.out, rsp...)
				p.pending--
				p.mu.Unlock()
			}()
		} else {
			p.out = append(p.out, rsp...)
			p.pending--
		}
	}
	return len(b), nil
}

// Read behaves like go.bug.st/serial: it returns whatever is available, or
// zero bytes and no error once the read timeout passes.
func (p *fakePak) Read(b []byte) (int, error) {
	p.mu.Lock()
	timeout := p.readTimeout
	p.mu.Unlock()
	if timeout <= 0 {
		timeout = time.Second
	}

	deadline := time.Now().Add(timeout)
	for {
		p.mu.Lock()
		if len(p.out) > 0 {
			n := copy(b, p.out)
			p.out = p.out[n:]
			p.mu.Unlock()
			return n, nil
		}
		p.mu.Unlock()
		if !time.Now().Before(deadline) {
			return 0, nil
		}
		time.Sleep(time.Millisecond)
	}
}

func (p *fakePak) stats() (commands int, violations []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.commands, append([]string(nil), p.violations...)
}

func newFakeDevice() (*Device, *fakePak) {
	p := &fakePak{}
	return &Device{f: &devicePort{Port: p}}, p
}

// wantCallerGaveUp checks that err is the error a caller gets when its own
// deadline passes: non-fatal, so SNI keeps the device open, and coded so the
// gRPC client sees DeadlineExceeded rather than Unknown.
func wantCallerGaveUp(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error; want a deadline exceeded error")
	}
	if devices.IsFatal(err) {
		t.Errorf("error is fatal: %v; a caller giving up must not make SNI close the device", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v; want it to wrap context.DeadlineExceeded", err)
	}
	var coded *devices.CodedError
	if !errors.As(err, &coded) || coded.Code != codes.DeadlineExceeded {
		t.Errorf("error = %v; want it to carry codes.DeadlineExceeded", err)
	}
}

func setHonorCallerDeadline(t *testing.T, v bool) {
	old := honorCallerDeadline
	honorCallerDeadline = v
	t.Cleanup(func() { honorCallerDeadline = old })
}

// The caller whose deadline passes mid-command gets its error at the deadline,
// the command is completed in the background, and the next request waits for
// that before touching the device -- then receives its own response, not the
// abandoned command's.
func TestTransact_callerDeadlineDrainsInBackground(t *testing.T) {
	setHonorCallerDeadline(t, true)
	d, pak := newFakeDevice()

	gate := make(chan struct{})
	pak.gate = gate

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, _, _, err := d.info(ctx)
	elapsed := time.Since(start)
	wantCallerGaveUp(t, err)
	if elapsed > 500*time.Millisecond {
		t.Errorf("caller got its error after %v; want it at its 50ms deadline", elapsed)
	}

	// The next request must queue behind the drain rather than writing its
	// command while the device still owes the abandoned one a response.
	type infoResult struct {
		rom string
		err error
	}
	next := make(chan infoResult, 1)
	go func() {
		_, _, rom, err := d.info(context.Background())
		next <- infoResult{rom, err}
	}()

	select {
	case r := <-next:
		t.Fatalf("next request completed (%q, %v) while the abandoned command was still draining", r.rom, r.err)
	case <-time.After(200 * time.Millisecond):
	}
	if n, _ := pak.stats(); n != 1 {
		t.Fatalf("device received %d commands while the first was draining; want 1", n)
	}

	close(gate)

	select {
	case r := <-next:
		if r.err != nil {
			t.Fatalf("next request failed after the drain: %v", r.err)
		}
		if r.rom != "seq1" {
			t.Errorf("next request got rom %q; want %q (its own response, not the abandoned one's)", r.rom, "seq1")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("next request did not complete after the drain")
	}

	if _, v := pak.stats(); len(v) > 0 {
		t.Errorf("protocol violations: %v", v)
	}
}

// A queued request whose own deadline passes while it waits for a drain fails
// with its own deadline error and never sends anything to the device.
func TestTransact_queuedCallerDeadlineWhileDraining(t *testing.T) {
	setHonorCallerDeadline(t, true)
	d, pak := newFakeDevice()

	gate := make(chan struct{})
	pak.gate = gate

	ctx1, cancel1 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel1()
	_, _, _, err := d.info(ctx1)
	wantCallerGaveUp(t, err)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	start := time.Now()
	_, _, _, err = d.info(ctx2)
	elapsed := time.Since(start)
	wantCallerGaveUp(t, err)
	if elapsed > 500*time.Millisecond {
		t.Errorf("queued caller got its error after %v; want it at its 100ms deadline", elapsed)
	}
	if n, _ := pak.stats(); n != 1 {
		t.Errorf("device received %d commands; the queued caller must not have sent one", n)
	}

	close(gate)

	_, _, rom, err := d.info(context.Background())
	if err != nil {
		t.Fatalf("request after the drain failed: %v", err)
	}
	if rom != "seq1" {
		t.Errorf("request after the drain got rom %q; want %q", rom, "seq1")
	}
	if _, v := pak.stats(); len(v) > 0 {
		t.Errorf("protocol violations: %v", v)
	}
}

// A caller whose deadline has already passed does not start a command at all.
func TestTransact_expiredCallerSendsNothing(t *testing.T) {
	setHonorCallerDeadline(t, true)
	d, pak := newFakeDevice()

	ctx, cancel := contextExpired()
	defer cancel()

	_, _, _, err := d.info(ctx)
	wantCallerGaveUp(t, err)
	if n, _ := pak.stats(); n != 0 {
		t.Errorf("device received %d commands from an already-expired caller; want 0", n)
	}

	// and the device is still perfectly usable:
	if _, _, rom, err := d.info(context.Background()); err != nil || rom != "seq0" {
		t.Errorf("info() after an expired caller = (%q, %v); want (\"seq0\", nil)", rom, err)
	}
}

// With the deadline policy off, the caller waits for the command to complete.
func TestTransact_honorCallerDeadlineOff(t *testing.T) {
	setHonorCallerDeadline(t, false)
	d, pak := newFakeDevice()

	gate := make(chan struct{})
	pak.gate = gate
	time.AfterFunc(200*time.Millisecond, func() { close(gate) })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, _, rom, err := d.info(ctx)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("info() with the deadline policy off failed after %v: %v", elapsed, err)
	}
	if rom != "seq0" {
		t.Errorf("rom = %q; want %q", rom, "seq0")
	}
	if elapsed < 150*time.Millisecond {
		t.Errorf("info() returned after %v; with the policy off it should have waited for the device", elapsed)
	}
}

// If an abandoned command fails to finish draining, the stream is in an
// unknown state, so the next request must report a fatal error to get the
// device reopened rather than talk to it.
func TestTransact_drainFailureIsFatalForNextRequest(t *testing.T) {
	setHonorCallerDeadline(t, true)
	d, _ := newFakeDevice()

	release := make(chan struct{})
	drained := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := transact(d, ctx, "test", func(ctx context.Context, _ *txn) (struct{}, error) {
		defer close(drained)
		<-release
		return struct{}{}, d.FatalError(errors.New("device stopped responding mid-drain"))
	})
	wantCallerGaveUp(t, err)

	close(release)
	<-drained

	ran := false
	_, err = transact(d, context.Background(), "next", func(ctx context.Context, _ *txn) (struct{}, error) {
		ran = true
		return struct{}{}, nil
	})
	if err == nil || !devices.IsFatal(err) {
		t.Errorf("next request error = %v; want a fatal error after a failed drain", err)
	}
	if ran {
		t.Errorf("next request talked to a device whose stream is out of step")
	}
}

// Once the caller has given up, data still arriving from the device goes to a
// null sink rather than into the caller's writer or progress callback.
func TestTransact_abandonedDataGoesToNullSink(t *testing.T) {
	setHonorCallerDeadline(t, true)
	d, _ := newFakeDevice()

	var buf bytes.Buffer
	var reports []uint32
	progress := func(current, total uint32) { reports = append(reports, current) }

	firstWritten := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan int, 1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-firstWritten
		cancel()
	}()

	_, err := transact(d, ctx, "test", func(ctx context.Context, t *txn) (struct{}, error) {
		w := t.writer(&buf)
		report := t.progress(progress)

		_, _ = w.Write([]byte("before"))
		report(6, 11)
		close(firstWritten)

		<-release
		n, _ := w.Write([]byte("after"))
		report(11, 11)
		finished <- n
		return struct{}{}, nil
	})
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("transact() error = %v; want context.Canceled", err)
	}
	var coded *devices.CodedError
	if !errors.As(err, &coded) || coded.Code != codes.Canceled {
		t.Errorf("error = %v; want it to carry codes.Canceled", err)
	}

	close(release)
	if n := <-finished; n != len("after") {
		t.Errorf("null sink reported %d bytes written; want %d so the drain carries on", n, len("after"))
	}

	if got := buf.String(); got != "before" {
		t.Errorf("caller's writer holds %q; want only %q", got, "before")
	}
	if len(reports) != 1 {
		t.Errorf("progress called %d times; want 1 (none after the caller gave up)", len(reports))
	}
}

// Once the caller has given up, the remainder of a payload the device is still
// owed is zero-filled rather than read from the caller's reader.
func TestTransact_abandonedReaderIsZeroFilled(t *testing.T) {
	setHonorCallerDeadline(t, true)
	d, _ := newFakeDevice()

	src := bytes.NewReader(bytes.Repeat([]byte{0xAA}, 8))
	firstRead := make(chan struct{})
	release := make(chan struct{})
	got := make(chan []byte, 1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-firstRead
		cancel()
	}()

	_, err := transact(d, ctx, "test", func(ctx context.Context, t *txn) (struct{}, error) {
		r := t.reader(src)
		b := make([]byte, 4)
		_, _ = r.Read(b)
		close(firstRead)

		<-release
		rest := bytes.Repeat([]byte{0xFF}, 4)
		_, _ = r.Read(rest)
		got <- rest
		return struct{}{}, nil
	})
	if err == nil {
		t.Fatalf("transact() returned nil error; want the caller to have given up")
	}

	close(release)
	if rest := <-got; !bytes.Equal(rest, make([]byte, 4)) {
		t.Errorf("read after the caller gave up = % x; want zeros", rest)
	}
	if src.Len() != 4 {
		t.Errorf("caller's reader has %d bytes left; want 4 (untouched after it gave up)", src.Len())
	}
}
