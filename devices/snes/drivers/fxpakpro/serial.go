package fxpakpro

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"go.bug.st/serial"
	"io"
	"sni/devices"
	"time"
)

const safeTimeout = time.Second * 1

type hasSetReadTimeout interface {
	// SetReadTimeout sets the timeout for the Read operation or use serial.NoTimeout
	// to disable read timeout.
	SetReadTimeout(t time.Duration) error
}

func readExact(ctx context.Context, f io.Reader, chunkSize uint32, buf []byte) (p uint32, err error) {
	// determine a deadline from context or default:
	var ok bool

	var deadline time.Time
	if deadline, ok = ctx.Deadline(); !ok {
		deadline = time.Now().Add(safeTimeout)
	}

	p = 0
	for p < chunkSize {
		// update the read timeout if applicable:
		if fr, ok := f.(hasSetReadTimeout); ok {
			timeout := deadline.Sub(time.Now())
			if timeout < 0 {
				// deadline already exceeded so cause Read() to fail instantly:
				timeout = 0
			}

			err = fr.SetReadTimeout(timeout)
			if err != nil {
				err = fmt.Errorf("readExact: setReadTimeout returned %w", err)
				return
			}
		}

		var n int
		n, err = f.Read(buf[p:chunkSize])
		if n < 0 {
			n = 0
		}
		if debugLog != nil {
			debugLog.Printf("readExact: read returned n=%d, err=%v\n%s", n, err, hex.Dump(buf[p:p+uint32(n)]))
		}
		p += uint32(n)
		if err != nil {
			return
		}
	}

	return
}

func writeExact(ctx context.Context, w io.Writer, chunkSize uint32, buf []byte) (p uint32, err error) {
	_ = ctx
	p = uint32(0)
	for p < chunkSize {
		var n int
		n, err = w.Write(buf[p:chunkSize])
		if n < 0 {
			n = 0
		}
		if debugLog != nil {
			debugLog.Printf("writeExact: write returned n=%d, err=%v\n%s", n, err, hex.Dump(buf[p:p+uint32(n)]))
		}
		p += uint32(n)
		if err != nil {
			return
		}
	}

	return
}

func sendSerial(f serial.Port, chunkSize uint32, buf []byte) (err error) {
	_, err = sendSerialProgress(f, chunkSize, uint32(len(buf)), bytes.NewReader(buf), nil)
	return
}

func sendSerialProgress(f serial.Port, chunkSize uint32, size uint32, r io.Reader, report devices.ProgressReportFunc) (sent uint32, err error) {
	// chunkSize is how many bytes each chunk is expected to be sized according to the protocol; valid values are [64, 512].
	if chunkSize != 64 && chunkSize != 512 {
		panic("chunkSize must be either 64 or 512")
	}

	var buf [512]byte

	ctx := context.Background()

	// transfer main chunks:
	chunks := size / chunkSize
	for i := uint32(0); i < chunks; i++ {
		if report != nil {
			report(sent, size)
		}

		var n uint32
		n, err = readExact(ctx, r, chunkSize, buf[:chunkSize])
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
	}

	// transfer any remainder:
	if size%chunkSize > 0 {
		if report != nil {
			report(sent, size)
		}

		var n uint32
		n, err = readExact(ctx, r, chunkSize, buf[:chunkSize])
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
