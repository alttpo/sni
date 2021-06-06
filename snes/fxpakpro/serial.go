package fxpakpro

import (
	"fmt"
	"go.bug.st/serial"
)

func sendSerial(f serial.Port, buf []byte) error {
	sent := 0
	for sent < len(buf) {
		n, e := f.Write(buf[sent:])
		if e != nil {
			return e
		}
		sent += n
	}
	return nil
}

func sendSerialProgress(f serial.Port, buf []byte, batchSize int, report func(sent int, total int)) error {
	sent := 0
	total := len(buf)
	for sent < total {
		report(sent, total)
		end := sent + batchSize
		if end > total {
			end = total
		}
		n, e := f.Write(buf[sent:end])
		if e != nil {
			return e
		}
		sent += n
	}
	if sent > total {
		sent = total
	}
	report(sent, total)
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
		o += n
	}
	return nil
}
