package fxpakpro

import (
	"fmt"
	"sni/snes"
)

type putfile struct {
	path   string
	rom    []byte
	report func(sent int, total int)
}

func newPUTFile(path string, rom []byte, report func(int, int)) *putfile {
	return &putfile{
		path,
		rom,
		report,
	}
}

func (c *putfile) Execute(queue snes.Queue, keepAlive snes.KeepAlive) error {
	f := queue.(*Queue).f

	sb := make([]byte, 512)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(c.path)
	copy(sb[256:512], nameBytes)

	// size of ROM contents:
	size := uint32(len(c.rom))
	sb[252] = byte((size >> 24) & 0xFF)
	sb[253] = byte((size >> 16) & 0xFF)
	sb[254] = byte((size >> 8) & 0xFF)
	sb[255] = byte((size >> 0) & 0xFF)

	// send command:
	err := sendSerial(f, sb)
	if err != nil {
		return err
	}

	// send data:
	err = sendSerialProgress(f, c.rom, 65536, func(sent int, total int) {
		// keep our command alive while we send data:
		keepAlive <- struct{}{}
		// report on progress:
		if c.report != nil {
			c.report(sent, total)
		}
	})
	if err != nil {
		return err
	}

	remainder := size & 511
	if remainder > 0 {
		// send however many 00 bytes that rounds up the size to the next 512 bytes:
		zeroes := make([]byte, 512-remainder)
		err = sendSerial(f, zeroes)
		if err != nil {
			return err
		}
	}

	// read response:
	rsp := make([]byte, 512)
	err = recvSerial(f, rsp, 512)
	if err != nil {
		return err
	}
	if rsp[0] != 'U' || rsp[1] != 'S' || rsp[2] != 'B' || rsp[3] != 'A' {
		return fmt.Errorf("putfile: %w", ErrInvalidResponse)
	}

	ec := rsp[5]
	if ec != 0 {
		return fmt.Errorf("putfile: error %d", ec)
	}

	return nil
}
