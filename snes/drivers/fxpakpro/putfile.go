package fxpakpro

import "fmt"

type putFileRequest struct {
	path   string
	rom    []byte
	report func(sent int, total int)
}

func (d *Device) putFile(doLock bool, req putFileRequest) (err error) {
	sb := make([]byte, 512)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// copy in the name to position 256:
	nameBytes := []byte(req.path)
	copy(sb[256:512], nameBytes)

	// size of ROM contents:
	size := uint32(len(req.rom))
	sb[252] = byte((size >> 24) & 0xFF)
	sb[253] = byte((size >> 16) & 0xFF)
	sb[254] = byte((size >> 8) & 0xFF)
	sb[255] = byte((size >> 0) & 0xFF)

	if doLock {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	// send command:
	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	// send data:
	err = sendSerialProgress(d.f, req.rom, 65536, func(sent int, total int) {
		// report on progress:
		if req.report != nil {
			req.report(sent, total)
		}
	})
	if err != nil {
		_ = d.Close()
		return
	}

	remainder := size & 511
	if remainder > 0 {
		// send however many 00 bytes that rounds up the size to the next 512 bytes:
		zeroes := make([]byte, 512-remainder)
		err = sendSerial(d.f, zeroes)
		if err != nil {
			_ = d.Close()
			return
		}
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
		return fmt.Errorf("putfile: unexpected response packet does not contain USBA header")
	}

	ec := rsp[5]
	if ec != 0 {
		return fmt.Errorf("putfile: fxpakpro responded with error code %d", ec)
	}

	return
}
