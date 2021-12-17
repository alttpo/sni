package fxpakpro

import (
	"bytes"
	"context"
	"fmt"
	"go.bug.st/serial"
	"io"
	"sni/snes"
	"time"
)

const safeTimeout = time.Second * 15

func sendSerial(f serial.Port, chunkSize int, buf []byte) (err error) {
	_, err = sendSerialProgress(f, chunkSize, uint32(len(buf)), bytes.NewReader(buf), nil)
	return
}

func sendSerialProgress(f serial.Port, chunkSize int, size uint32, r io.Reader, report snes.ProgressReportFunc) (sent uint32, err error) {
	// chunkSize is how many bytes each chunk is expected to be sized according to the protocol; valid values are [64, 512].
	if chunkSize != 64 && chunkSize != 512 {
		panic("chunkSize must be either 64 or 512")
	}

	buf := make([]byte, chunkSize)

	sent = uint32(0)
	for sent < size {
		if report != nil {
			report(sent, size)
		}

		var nr int
		nr, err = r.Read(buf)
		if err != nil {
			err = fmt.Errorf("sendSerialProgress: read from io.Reader: %w", err)
			return
		}
		if nr != chunkSize {
			// should be last chunk; zero out the remaining bytes:
			for i := nr; i < chunkSize; i++ {
				buf[i] = 0
			}
			nr = chunkSize
		}

		// write to serial port in chunkSize bytes every time:
		var nw int
		nw, err = f.Write(buf)
		if err != nil {
			return
		}
		if nw != nr {
			err = fmt.Errorf("sendSerialProgress: read %d bytes but only write %d to serial port", nr, nw)
			return
		}

		sent += uint32(nw)
	}
	if sent > size {
		sent = size
	}

	if report != nil {
		report(sent, size)
	}
	return
}

func readExact(ctx context.Context, f serial.Port, chunkSize int, buf []byte) (err error) {
	// determine a deadline from context or default:
	var ok bool
	var deadline time.Time
	if deadline, ok = ctx.Deadline(); !ok {
		deadline = time.Now().Add(safeTimeout)
	}

	p := 0
	for p < chunkSize {
		timeout := deadline.Sub(time.Now())
		if timeout < 0 {
			// deadline already exceeded so cause Read() to fail instantly:
			timeout = 0
		}
		err = f.SetReadTimeout(timeout)
		if err != nil {
			err = fmt.Errorf("readExact: setReadTimeout returned %w", err)
			return
		}

		var n int
		n, err = f.Read(buf[p:chunkSize])
		if err != nil {
			return
		}
		if n <= 0 {
			err = fmt.Errorf("readExact: read returned %d", n)
			return
		}
		p += n
	}

	return
}

func recvSerial(ctx context.Context, f serial.Port, rsp []byte, expected int) (err error) {
	err = readExact(ctx, f, expected, rsp)
	if err != nil {
		err = fmt.Errorf("recvSerial: %w", err)
		return
	}
	return
}

func recvSerialProgress(ctx context.Context, f serial.Port, w io.Writer, size uint32, chunkSize int, progress snes.ProgressReportFunc) (received uint32, err error) {
	buf := make([]byte, chunkSize)

	received = uint32(0)
	for received < size {
		if progress != nil {
			progress(received, size)
		}

		err = readExact(ctx, f, chunkSize, buf)
		if err != nil {
			err = fmt.Errorf("recvSerialProgress: %w", err)
			return
		}

		received += uint32(chunkSize)
		if received <= size {
			_, err = w.Write(buf)
			if err != nil {
				return
			}
		} else {
			_, err = w.Write(buf[0 : received-size])
			if err != nil {
				return
			}
			received = size
		}
	}

	if progress != nil {
		progress(received, size)
	}

	err = f.SetReadTimeout(serial.NoTimeout)
	if err != nil {
		err = fmt.Errorf("recvSerialProgress: %w", err)
		return
	}

	return
}
