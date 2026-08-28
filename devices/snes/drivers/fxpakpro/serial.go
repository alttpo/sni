package fxpakpro

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"go.bug.st/serial"
	"io"
	"log"
	"runtime/trace"
	"sni/devices"
	"time"
)

const safeTimeout = time.Second * 1

// Timeout and cancellation policy for talking to the device. These are vars
// rather than consts because DriverInit() overrides them from configuration
// (SNI_FXPAKPRO_READ_TIMEOUT, SNI_FXPAKPRO_WRITE_TIMEOUT and
// SNI_FXPAKPRO_HONOR_CALLER_DEADLINE); tests also adjust them directly. The
// values here are the defaults used when configuration has not been loaded.
var (
	// noDataTimeout is how long readExact waits with no bytes at all arriving
	// before declaring the device unresponsive.
	noDataTimeout = time.Second * 15

	// writeTimeout is how long a single write is allowed to make no progress
	// before the device is declared unable to accept data. See
	// writeWithTimeout.
	writeTimeout = time.Second * 15

	// honorCallerDeadline controls whether a caller's context deadline or
	// cancellation aborts I/O already in flight to the device. When false, only
	// the two timeouts above bound device I/O, so a transfer that has already
	// started is not cut short by an impatient client. The device has no way to
	// be told that a command was abandoned, so stopping midway through one
	// leaves the protocol out of step until the stream is drained.
	honorCallerDeadline = true
)

func readExactGeneric(ctx context.Context, f io.Reader, chunkSize uint32, buf []byte) (p uint32, err error) {
	ctx, task := trace.NewTask(ctx, "readExactGeneric")
	defer task.End()

	attempts := 0
	p = 0
	for p < chunkSize {
		var n int
		lastp := p
		n, err = f.Read(buf[p:chunkSize])
		trace.Logf(ctx, "read", "read(buf[%d:%d]) = %v, %v", p, chunkSize, n, err)
		if n < 0 {
			n = 0
		}
		p += uint32(n)
		if p == lastp {
			attempts++
			trace.Logf(ctx, "retry", "attempts = %v", attempts)
			if attempts >= 60 {
				err = fmt.Errorf("readExactGeneric: timed out after 60 attempts of reading zero bytes")
				return
			}
		} else {
			attempts = 0
		}
		if err != nil {
			return
		}
	}

	trace.Log(ctx, "read", "return")
	return
}

func readExact(ctx context.Context, f serial.Port, chunkSize uint32, buf []byte) (p uint32, err error) {
	// determine a deadline from context or default:
	var ok bool

	ctx, task := trace.NewTask(ctx, "readExact")
	defer task.End()

	haveHardDeadline := false
	var deadline time.Time
	if honorCallerDeadline {
		if deadline, ok = ctx.Deadline(); ok {
			trace.Logf(ctx, "deadline", "deadline=%v", deadline)
			haveHardDeadline = true
		}
	}

	// Budget for how long the device may stay completely silent before we give
	// up. This is a total elapsed time since the last byte arrived, not a count
	// of read attempts. FatFs runs inside the firmware's USB interrupt handler,
	// and a cluster allocation on a large, full, fragmented card can scan a
	// sizeable fraction of the FAT before the device can reply: a 64 GB card
	// with 32 KB clusters has an ~8 MB FAT, so a worst-case scan is on the order
	// of ten seconds.
	lastProgress := time.Now()
	p = 0
	for p < chunkSize {
		// update the read timeout if applicable:
		{
			var timeout time.Duration
			if haveHardDeadline {
				// we have a hard deadline to meet:
				timeout = time.Until(deadline)
				if timeout <= 0 {
					// Deadline already exceeded. Return now rather than looping:
					// Read() would come straight back with zero bytes and spin
					// until the silence budget expired.
					err = fmt.Errorf(
						"readExact: context deadline exceeded after reading %d of %d bytes",
						p, chunkSize)
					return
				}
			} else {
				// no hard deadline; each read() attempt gets its own timeout:
				timeout = safeTimeout
			}

			trace.Logf(ctx, "deadline", "SetReadTimeout(%v)", timeout)
			err = f.SetReadTimeout(timeout)
			if err != nil {
				err = fmt.Errorf("readExact: setReadTimeout returned %w", err)
				return
			}
		}

		var n int
		lastp := p
		n, err = f.Read(buf[p:chunkSize])
		trace.Logf(ctx, "read", "read(buf[%d:%d]) = %v, %v", p, chunkSize, n, err)
		if n < 0 {
			n = 0
		}
		if debugLog != nil {
			debugLog.Printf("readExact: read returned n=%d, err=%v\n%s", n, err, hex.Dump(buf[p:p+uint32(n)]))
		}
		p += uint32(n)
		if p == lastp {
			silent := time.Since(lastProgress)
			trace.Logf(ctx, "retry", "silent for %v", silent)
			if silent >= noDataTimeout {
				err = fmt.Errorf(
					"readExact: no data from device for %v after reading %d of %d bytes",
					silent, p, chunkSize)
				return
			}
		} else {
			lastProgress = time.Now()
		}
		if err != nil {
			return
		}
	}

	trace.Log(ctx, "read", "return")
	return
}

// blockingWrite writes all of buf to w. It can block indefinitely, so it is
// only ever called on a goroutine owned by writeWithTimeout.
func blockingWrite(w io.Writer, buf []byte) (p uint32, err error) {
	for p < uint32(len(buf)) {
		var n int
		n, err = w.Write(buf[p:])
		if n < 0 {
			n = 0
		}
		if debugLog != nil {
			debugLog.Printf("write returned n=%d, err=%v\n%s", n, err, hex.Dump(buf[p:p+uint32(n)]))
		}
		if err != nil {
			return
		}
		p += uint32(n)
	}
	return
}

// writeWithTimeout writes all of buf to w, giving up if the device stops
// accepting data.
//
// This has to be done on a separate goroutine because there is no way to bound
// the write itself. go.bug.st/serial's Port interface exposes SetReadTimeout
// but no SetWriteTimeout, and on Windows the library sets
// WriteTotalTimeoutConstant and WriteTotalTimeoutMultiplier to 0, which Win32
// COMMTIMEOUTS defines as "wait forever". So when the fxpakpro stops draining
// its USB OUT endpoint and NAKs indefinitely, a plain Write() never returns.
// Left unbounded that hangs the calling goroutine while it holds d.lock, which
// blocks every other request for that device -- SNI appears to freeze rather
// than reporting a failed transfer.
//
// When we give up, the goroutine is still inside Write() and may yet put bytes
// on the wire. It must never be left running alongside a later command: the
// caller releases d.lock on the way out, so the next command would write
// concurrently with the orphan and interleave with it, corrupting the protocol.
// (Observed in practice: after an abandoned write the following command read
// the abandoned one's USBA response.)
//
// So abandoning a write also closes the port. That makes the pending Write()
// fail so the goroutine exits, and guarantees every later write on this port
// fails immediately instead of racing. Both callers already treat this error as
// fatal, which makes SNI close and reopen the device anyway; doing it here just
// removes the window in between.
func writeWithTimeout(ctx context.Context, w io.Writer, buf []byte) (p uint32, err error) {
	type writeResult struct {
		p   uint32
		err error
	}
	// buffered so the goroutine can always finish even if we stopped waiting:
	done := make(chan writeResult, 1)
	go func() {
		wp, werr := blockingWrite(w, buf)
		done <- writeResult{wp, werr}
	}()

	// Wait on the caller's context and our own budget separately, rather than
	// clamping one to the other, so the error names the actual cause: a
	// cancelled request and an unresponsive device are different problems.
	//
	// When the caller's deadline is not honored, cancelled stays nil, and a
	// receive on a nil channel blocks forever -- which disables that arm of the
	// select without needing a second copy of it.
	var cancelled <-chan struct{}
	if honorCallerDeadline {
		cancelled = ctx.Done()
	}

	timer := time.NewTimer(writeTimeout)
	defer timer.Stop()

	select {
	case r := <-done:
		trace.Logf(ctx, "write", "write(%d bytes) = %v, %v", len(buf), r.p, r.err)
		return r.p, r.err
	case <-cancelled:
		abandonPort(w)
		return 0, fmt.Errorf("write: abandoned %d byte write: %w", len(buf), ctx.Err())
	case <-timer.C:
		abandonPort(w)
		// how much made it out is unknown, so report none of it:
		return 0, fmt.Errorf(
			"write: device accepted no data for %v while writing %d bytes; "+
				"it is not draining its USB endpoint", writeTimeout, len(buf))
	}
}

// abandonPort closes the port underneath a write we have stopped waiting for,
// so the orphaned goroutine cannot interleave with whatever the caller does
// next. Closing is what unblocks it: a pending Write() fails once the handle is
// gone. Errors are logged rather than returned; the caller already has a more
// useful error describing why the write was abandoned.
func abandonPort(w io.Writer) {
	c, ok := w.(io.Closer)
	if !ok {
		return
	}
	if err := c.Close(); err != nil {
		log.Printf("%s: closing port after an abandoned write: %v\n", driverName, err)
	}
}

func writeExact(ctx context.Context, w io.Writer, chunkSize uint32, buf []byte) (p uint32, err error) {
	ctx, task := trace.NewTask(ctx, "writeExact")
	defer task.End()

	return writeWithTimeout(ctx, w, buf[:chunkSize])
}

func sendSerial(ctx context.Context, f serial.Port, buf []byte) (err error) {
	ctx, task := trace.NewTask(ctx, "sendSerial")
	defer task.End()

	_, err = writeWithTimeout(ctx, f, buf)
	return
}

func sendSerialChunked(ctx context.Context, f serial.Port, chunkSize uint32, buf []byte) (err error) {
	_, err = sendSerialProgress(ctx, f, chunkSize, uint32(len(buf)), bytes.NewReader(buf), nil)
	return
}

func sendSerialProgress(ctx context.Context, f serial.Port, chunkSize uint32, size uint32, r io.Reader, report devices.ProgressReportFunc) (sent uint32, err error) {
	// chunkSize is how many bytes each chunk is expected to be sized according to the protocol; valid values are [64, 512].
	if chunkSize != 64 && chunkSize != 512 {
		panic("chunkSize must be either 64 or 512")
	}

	var buf [512]byte

	// transfer main chunks:
	chunks := size / chunkSize
	for i := uint32(0); i < chunks; i++ {
		if report != nil {
			report(sent, size)
		}

		// read from the upstream reader:
		var n uint32
		n, err = readExactGeneric(ctx, r, chunkSize, buf[:chunkSize])
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			return
		}

		// zero out remaining bytes of chunk, if any:
		for ; n < chunkSize; n++ {
			buf[n] = 0
		}

		n, err = writeExact(ctx, f, chunkSize, buf[:chunkSize])
		sent += n
		if err != nil {
			// bail out immediately; continuing would silently drop this error
			// when the next iteration reassigns err, reporting a short or
			// corrupt transfer as a success:
			err = fmt.Errorf("sendSerialProgress: write failed after %d of %d bytes: %w", sent, size, err)
			return
		}
	}

	// transfer any remainder:
	if size%chunkSize > 0 {
		if report != nil {
			report(sent, size)
		}

		var n uint32
		n, err = readExactGeneric(ctx, r, chunkSize, buf[:chunkSize])
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			return
		}

		// zero out remaining bytes of chunk, if any:
		for ; n < chunkSize; n++ {
			buf[n] = 0
		}

		n, err = writeExact(ctx, f, chunkSize, buf[:chunkSize])
		sent += n
		if err != nil {
			err = fmt.Errorf("sendSerialProgress: write failed after %d of %d bytes: %w", sent, size, err)
			return
		}
	}

	// final progress report:
	if report != nil {
		report(sent, size)
	}

	return
}

func recvSerial(ctx context.Context, f serial.Port, rsp []byte, expected uint32) (err error) {
	_, err = readExact(ctx, f, expected, rsp)
	if err != nil {
		err = fmt.Errorf("recvSerial: %w", err)
		return
	}
	return
}
