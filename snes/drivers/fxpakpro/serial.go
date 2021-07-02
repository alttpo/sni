package fxpakpro

import (
	"fmt"
	"go.bug.st/serial"
	"io"
	"sni/snes"
)

func sendSerial(f serial.Port, chunkSize int, buf []byte) error {
	return sendSerialProgress(f, chunkSize, buf, nil)
}

func sendSerialProgress(f serial.Port, chunkSize int, buf []byte, report snes.ProgressReportFunc) error {
	// chunkSize is how many bytes each chunk is expected to be sized according to the protocol; valid values are [64, 512].
	if chunkSize != 64 && chunkSize != 512 {
		panic("chunkSize must be either 64 or 512")
	}

	sent := uint64(0)
	total := uint64(len(buf))
	for sent < total {
		if report != nil {
			report(sent, total)
		}
		end := sent + uint64(chunkSize)
		if end > total {
			end = total
		}
		// always make sure we're sending chunks of equivalent size:
		chunk := buf[sent:end]
		if len(chunk) < chunkSize {
			chunk = make([]byte, chunkSize)
			copy(chunk, buf[sent:end])
		}
		n, e := f.Write(chunk)
		if e != nil {
			return e
		}
		sent += uint64(n)
	}
	if sent > total {
		sent = total
	}
	if report != nil {
		report(sent, total)
	}
	return nil
}

func recvSerial(f serial.Port, rsp []byte, expected int) error {
	o := 0
	for o < expected {
		n, err := f.Read(rsp[o:expected])
		if err != nil {
			return err
		}
		if n <= 0 {
			return fmt.Errorf("recvSerial: Read returned %d", n)
		}

		//log.Printf("<< [%d:%d]\n%s", o, o+n, hex.Dump(rsp[o:o+n]))
		o += n
	}
	return nil
}

func recvSerialProgress(f serial.Port, w io.Writer, expected uint64, chunkSize int, progress snes.ProgressReportFunc) (received uint64, err error) {
	buf := make([]byte, chunkSize)

	received = uint64(0)
	for received < expected {
		p := 0
		for p < chunkSize {
			var n int
			n, err = f.Read(buf[p:chunkSize])
			if err != nil {
				return
			}
			if n <= 0 {
				return received, fmt.Errorf("recvSerialProgress: Read returned %d", n)
			}
			p += n
		}

		received += uint64(chunkSize)
		if received <= expected {
			_, err = w.Write(buf)
			if err != nil {
				return
			}
		} else {
			_, err = w.Write(buf[0 : received-expected])
			if err != nil {
				return
			}
			received = expected
		}

		if progress != nil {
			progress(received, expected)
		}
	}

	return
}
