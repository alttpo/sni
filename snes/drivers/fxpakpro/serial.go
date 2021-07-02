package fxpakpro

import (
	"bytes"
	"fmt"
	"go.bug.st/serial"
	"io"
	"sni/snes"
)

func readExact(r io.Reader, chunkSize int, buf []byte) (err error) {
	p := 0
	for p < chunkSize {
		var n int
		n, err = r.Read(buf[p:chunkSize])
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

		var n int
		n, err = r.Read(buf)
		if err != nil {
			err = fmt.Errorf("sendSerialProgress: %w", err)
			return
		}

		// write to serial port:
		n, err = f.Write(buf)
		if err != nil {
			return
		}

		sent += uint32(n)
	}
	if sent > size {
		sent = size
	}

	if report != nil {
		report(sent, size)
	}
	return
}

func recvSerial(f serial.Port, rsp []byte, expected int) (err error) {
	err = readExact(f, expected, rsp)
	if err != nil {
		err = fmt.Errorf("recvSerial: %w", err)
		return
	}
	return
}

func recvSerialProgress(f serial.Port, w io.Writer, size uint32, chunkSize int, progress snes.ProgressReportFunc) (received uint32, err error) {
	buf := make([]byte, chunkSize)

	received = uint32(0)
	for received < size {
		if progress != nil {
			progress(received, size)
		}

		err = readExact(f, chunkSize, buf)
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

	return
}
