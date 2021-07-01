package fxpakpro

import (
	"fmt"
	"go.bug.st/serial"
)

func sendSerial(f serial.Port, chunkSize int, buf []byte) error {
	return sendSerialProgress(f, chunkSize, buf, nil)
}

func sendSerialProgress(f serial.Port, chunkSize int, buf []byte, report func(sent int, total int)) error {
	// chunkSize is how many bytes each chunk is expected to be sized according to the protocol; valid values are [64, 512].
	if chunkSize != 64 && chunkSize != 512 {
		panic("chunkSize must be either 64 or 512")
	}

	sent := 0
	total := len(buf)
	for sent < total {
		if report != nil {
			report(sent, total)
		}
		end := sent + chunkSize
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
		sent += n
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
