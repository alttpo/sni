package fxpakpro

import (
	"fmt"
	"sni/snes"
)

type boot struct {
	path string
}

func newBOOT(path string) *boot {
	return &boot{path: path}
}

func (c *boot) Execute(queue snes.Queue, keepAlive snes.KeepAlive) error {
	f := queue.(*Queue).f

	sb := make([]byte, 512)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpBOOT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the path to position 256:
	nameBytes := []byte(c.path)
	copy(sb[256:512], nameBytes)

	// size isn't used for BOOT:
	size := uint32(0)
	sb[252] = byte((size >> 24) & 0xFF)
	sb[253] = byte((size >> 16) & 0xFF)
	sb[254] = byte((size >> 8) & 0xFF)
	sb[255] = byte((size >> 0) & 0xFF)

	// send command:
	err := sendSerial(f, sb)
	if err != nil {
		return err
	}

	// read response:
	rsp := make([]byte, 512)
	err = recvSerial(f, rsp, 512)
	if err != nil {
		return err
	}
	if rsp[0] != 'U' || rsp[1] != 'S' || rsp[2] != 'B' || rsp[3] != 'A' {
		return fmt.Errorf("boot: %w", ErrInvalidResponse)
	}

	//ec := rsp[5]
	//if ec != 0 {
	//	return fmt.Errorf("mkdir: error %d", ec)
	//}

	return nil
}
