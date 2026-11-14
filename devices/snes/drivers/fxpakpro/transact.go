package fxpakpro

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sni/devices"
	"sync"

	"google.golang.org/grpc/codes"
)

// The fxpakpro speaks a strictly synchronous request/response protocol and has
// no way to be told that a command was abandoned. Once a command has started,
// every byte the protocol calls for has to cross the wire in both directions,
// or the next command is framed against leftovers from this one. Closing and
// reopening the port does not reliably reset the firmware's own state either,
// so stopping partway through is never a safe way to honor a caller's deadline.
//
// So the caller's deadline and the device's I/O are decoupled:
//
//   - Each command/response exchange (a transaction) runs on its own goroutine
//     with a context that carries no caller cancellation. It always runs to the
//     end of the protocol exchange, bounded only by the driver's own device
//     timeouts (noDataTimeout, writeTimeout).
//
//   - The caller waits for that goroutine or for its own context, whichever
//     comes first. When the caller's deadline passes, it gets its error at that
//     moment and returns; the transaction carries on as a drain. Anything the
//     device still sends goes to a null sink rather than the caller's buffers,
//     writer or progress callbacks, and anything the protocol still requires us
//     to send is sent (zero-filled once the caller's reader can no longer be
//     touched).
//
//   - The transaction holds d.port until it has fully completed, drain
//     included. The next request queues on d.port and so does not touch the
//     device until the stream is back in step. If its own deadline passes while
//     it waits, it fails without ever having started a command.
//
//   - If a drain itself fails, the stream is in an unknown state. That is
//     recorded and the next request reports it as fatal, which makes SNI close
//     and reopen the device -- the same outcome as any other fatal I/O error.

// ctxMutex is a mutex whose Lock can give up when a context ends. The zero
// value is unlocked and ready to use.
type ctxMutex struct {
	once sync.Once
	ch   chan struct{}
}

func (m *ctxMutex) sem() chan struct{} {
	m.once.Do(func() { m.ch = make(chan struct{}, 1) })
	return m.ch
}

// Lock blocks until the mutex is acquired.
func (m *ctxMutex) Lock() { m.sem() <- struct{}{} }

// LockContext acquires the mutex, or returns ctx's error if the caller's
// deadline passes first. With honorCallerDeadline off it waits regardless.
func (m *ctxMutex) LockContext(ctx context.Context) error {
	sem := m.sem()
	if !honorCallerDeadline {
		sem <- struct{}{}
		return nil
	}
	// An already-expired caller must not take the lock even when it is free;
	// select would otherwise pick between the two ready cases at random.
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *ctxMutex) Unlock() {
	select {
	case <-m.sem():
	default:
		panic("fxpakpro: unlock of unlocked ctxMutex")
	}
}

// lockOp takes d.lock for the duration of a client operation, unless ctx shows
// the caller already holds it (a composite operation issuing several
// transactions). The returned context marks the lock as held.
func (d *Device) lockOp(ctx context.Context, op string) (context.Context, func(), error) {
	if !shouldLock(ctx) {
		return ctx, func() {}, nil
	}
	if err := d.lock.LockContext(ctx); err != nil {
		return ctx, nil, d.callerGaveUp(op, "while waiting for the device", err)
	}
	return context.WithValue(ctx, lockedKey, &struct{}{}), d.lock.Unlock, nil
}

// callerGaveUp builds the error returned to a caller whose context ended. It is
// non-fatal: the device itself is fine (or is being brought back into step in
// the background), so it must not be closed. It carries the matching gRPC code
// so the client sees DeadlineExceeded or Canceled rather than Unknown.
func (d *Device) callerGaveUp(op, when string, cause error) devices.DeviceError {
	code := codes.Canceled
	if errors.Is(cause, context.DeadlineExceeded) {
		code = codes.DeadlineExceeded
	}
	return d.NonFatalError(devices.WithCode(code, fmt.Errorf("%s: %s: %w", op, when, cause)))
}

// txn is the link between a transaction's I/O goroutine and the caller waiting
// on it. Everything the I/O goroutine hands to caller-owned state goes through
// deliver, so once the caller has detached nothing of theirs is touched again.
type txn struct {
	mu        sync.Mutex
	abandoned bool // the caller stopped waiting; deliver nothing further
	finished  bool // the I/O goroutine has a result for the caller
}

// deliver runs f, which touches caller-owned state, unless the caller has
// already detached. It reports whether f ran.
func (t *txn) deliver(f func()) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.abandoned {
		return false
	}
	f()
	return true
}

// writer returns w wrapped so that, once the caller detaches, data the device
// still sends is discarded instead of written to w.
func (t *txn) writer(w io.Writer) io.Writer {
	if w == nil {
		return nil
	}
	return txnWriter{t, w}
}

type txnWriter struct {
	t *txn
	w io.Writer
}

func (x txnWriter) Write(p []byte) (n int, err error) {
	if !x.t.deliver(func() { n, err = x.w.Write(p) }) {
		return len(p), nil // null sink
	}
	return
}

// reader returns r wrapped so that, once the caller detaches, the rest of the
// payload the protocol still requires is supplied as zeros rather than read
// from a reader the caller may already have discarded.
func (t *txn) reader(r io.Reader) io.Reader {
	if r == nil {
		return nil
	}
	return txnReader{t, r}
}

type txnReader struct {
	t *txn
	r io.Reader
}

func (x txnReader) Read(p []byte) (n int, err error) {
	if !x.t.deliver(func() { n, err = x.r.Read(p) }) {
		clear(p)
		return len(p), nil
	}
	return
}

// progress returns report wrapped so that it is no longer called once the
// caller detaches.
func (t *txn) progress(report devices.ProgressReportFunc) devices.ProgressReportFunc {
	if report == nil {
		return nil
	}
	return func(current, total uint32) {
		t.deliver(func() { report(current, total) })
	}
}

// transact runs fn, one complete command/response exchange with the device,
// as described at the top of this file. fn receives a context that carries the
// caller's values but not its deadline or cancellation, and must return its
// results rather than write them into caller-owned memory; anything it must
// hand over while still running (streamed file data, progress) goes through t.
//
// The caller must hold d.lock (see lockOp).
func transact[T any](d *Device, ctx context.Context, op string, fn func(ctx context.Context, t *txn) (T, error)) (res T, err error) {
	if err = d.port.LockContext(ctx); err != nil {
		err = d.callerGaveUp(op, "while waiting for an earlier abandoned command to finish", err)
		return
	}

	// d.drainErr is only touched while holding d.port.
	if d.drainErr != nil {
		derr := d.drainErr
		d.port.Unlock()
		err = d.FatalError(fmt.Errorf("%s: device stream is out of step after an abandoned command failed to finish: %w", op, derr))
		return
	}

	type result struct {
		v   T
		err error
	}
	t := &txn{}
	done := make(chan result, 1)
	go func() {
		v, ferr := fn(context.WithoutCancel(ctx), t)

		t.mu.Lock()
		abandoned := t.abandoned
		t.finished = true
		t.mu.Unlock()

		if abandoned {
			if ferr == nil {
				if debugLog != nil {
					debugLog.Printf("%s: abandoned command finished draining\n", op)
				}
			} else if devices.IsFatal(ferr) {
				log.Printf("%s: %s: abandoned command failed to finish draining: %v\n", driverName, op, ferr)
				d.drainErr = ferr
			}
		}

		// release the device to the next request only now the protocol exchange
		// is over, and before the caller hears about it:
		d.port.Unlock()
		done <- result{v, ferr}
	}()

	// With the deadline policy off, cancelled stays nil and only the device's
	// own timeouts bound how long the caller waits.
	var cancelled <-chan struct{}
	if honorCallerDeadline {
		cancelled = ctx.Done()
	}

	select {
	case r := <-done:
		return r.v, r.err
	case <-cancelled:
		t.mu.Lock()
		if t.finished {
			// the transaction completed as the deadline passed; its result is
			// already on its way, so report it rather than a spurious timeout:
			t.mu.Unlock()
			r := <-done
			return r.v, r.err
		}
		t.abandoned = true
		t.mu.Unlock()

		if debugLog != nil {
			debugLog.Printf("%s: caller gave up (%v); draining the command in the background\n", op, ctx.Err())
		}
		err = d.callerGaveUp(op, "abandoned mid-command; the device is being brought back into step in the background", ctx.Err())
		return
	}
}
