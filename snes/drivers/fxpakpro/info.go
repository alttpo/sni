package fxpakpro

import (
	"context"
	"fmt"
	"sni/snes"
)

func (d *Device) FetchFields(ctx context.Context, fields ...snes.Field) (values []string, err error) {
	var version string
	var deviceName string
	var rom string

	version, deviceName, rom, err = d.info(ctx)
	if err != nil {
		return
	}

	for _, field := range fields {
		switch field {
		case snes.Field_DeviceName:
			values = append(values, deviceName)
			break
		case snes.Field_DeviceVersion:
			values = append(values, version)
			break
		case snes.Field_RomFileName:
			values = append(values, rom)
			break
		default:
			// unknown value; append empty string to maintain index association:
			values = append(values, "")
			break
		}
	}

	return
}

func (d *Device) info(ctx context.Context) (version, device, rom string, err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpINFO)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	// send command:
	err = sendSerial(d.f, 512, sb)
	if err != nil {
		err = d.FatalError(err)
		_ = d.Close()
		return
	}

	// read response:
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		err = d.FatalError(err)
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		err = fmt.Errorf("info: fxpakpro response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if sb[4] != byte(OpRESPONSE) {
		_ = d.Close()
		err = fmt.Errorf("info: wrong opcode in response packet; got $%02x", sb[4])
		err = d.FatalError(err)
		return
	}
	if ec := sb[5]; ec != 0 {
		err = fmt.Errorf("info: %w", fxpakproError(ec))
		err = d.NonFatalError(err)
		return
	}

	romB := sb[16:252]
	rom = string(romB[:clen(romB)])

	versionB := sb[260 : 260+64]
	version = string(versionB[:clen(versionB)])

	deviceB := sb[260+64 : 260+64+64]
	device = string(deviceB[:clen(deviceB)])

	return
}

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
