package fxpakpro

import (
	"context"
	"fmt"
)

func (d *Device) mkdir(ctx context.Context, path string) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpMKDIR)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(path)
	copy(sb[256:512], nameBytes)

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	// send command:
	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	// read response:
	rsp := make([]byte, 512)
	err = recvSerial(d.f, rsp, 512)
	if err != nil {
		_ = d.Close()
		return
	}
	if rsp[0] != 'U' || rsp[1] != 'S' || rsp[2] != 'B' || rsp[3] != 'A' {
		_ = d.Close()
		return fmt.Errorf("mkdir: fxpakpro response packet does not contain USBA header")
	}

	if ec := rsp[5]; ec != 0 {
		return fmt.Errorf("mkdir: fxpakpro responded with error code %d", ec)
	}

	return
}
