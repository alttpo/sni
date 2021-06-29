package fxpakpro

import "encoding/binary"

func (d *Device) put(space space, address uint32, data []byte) (err error) {
	sb := make([]byte, 512)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(space)
	sb[6] = byte(FlagNONE)

	// put the data size in:
	size := uint32(len(data))
	binary.BigEndian.PutUint32(sb[252:], size)

	// put the address in:
	binary.BigEndian.PutUint32(sb[256:], address)

	// send the data to the USB port:
	defer d.lock.Unlock()
	d.lock.Lock()

	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	dest := sb[0:]
	for len(data) > 0 {
		var n int
		for i := range dest {
			dest[i] = 0
		}
		n = copy(dest, data)
		data = data[n:]

		err = sendSerial(d.f, sb)
		if err != nil {
			_ = d.Close()
			return
		}
	}

	// await single response:
	err = recvSerial(d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}

	return
}
