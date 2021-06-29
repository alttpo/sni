package fxpakpro

import "context"

func (d *Device) ResetSystem(ctx context.Context) (err error) {
	sb := make([]byte, 512)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpRESET)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	d.lock.Lock()
	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		d.lock.Unlock()
		return
	}

	err = recvSerial(d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		d.lock.Unlock()
		return
	}

	d.lock.Unlock()
	return
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	panic("implement me")
}

func (d *Device) PauseToggle(ctx context.Context) error {
	panic("implement me")
}
