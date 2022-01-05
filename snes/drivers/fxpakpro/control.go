package fxpakpro

import (
	"context"
	"fmt"
)

func (d *Device) ResetSystem(ctx context.Context) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpRESET)
	sb[5] = byte(SpaceSNES)
	sb[6] = byte(FlagNONE)

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	err = sendSerial(d.f, 512, sb)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		err = fmt.Errorf("reset: fxpakpro response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		err = fmt.Errorf("reset: %w", fxpakproError(ec))
		err = d.NonFatalError(err)
		return
	}

	return
}

func (d *Device) ResetToMenu(ctx context.Context) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpMENU_RESET)
	sb[5] = byte(SpaceSNES)
	sb[6] = byte(FlagNONE)

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	err = sendSerial(d.f, 512, sb)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		err = fmt.Errorf("menu_reset: fxpakpro response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		err = fmt.Errorf("menu_reset: %w", fxpakproError(ec))
		err = d.NonFatalError(err)
		return
	}

	return
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	panic("implement me")
}

func (d *Device) PauseToggle(ctx context.Context) error {
	panic("implement me")
}
