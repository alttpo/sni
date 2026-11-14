package fxpakpro

import (
	"context"
	"fmt"
	"sni/protos/sni"
)

func (d *Device) FetchFields(ctx context.Context, fields ...sni.Field) (values []string, err error) {
	var version string
	var deviceName string
	var rom string

	version, deviceName, rom, err = d.info(ctx)
	if err != nil {
		return
	}

	for _, field := range fields {
		switch field {
		case sni.Field_DeviceName:
			values = append(values, deviceName)
			break
		case sni.Field_DeviceVersion:
			values = append(values, version)
			break
		case sni.Field_RomFileName:
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

type infoResult struct {
	version, device, rom string
}

func (d *Device) info(ctx context.Context) (version, device, rom string, err error) {
	ctx, unlock, err := d.lockOp(ctx, "info")
	if err != nil {
		return
	}
	defer unlock()

	var r infoResult
	r, err = transact(d, ctx, "info", d.infoIO)
	return r.version, r.device, r.rom, err
}

// infoIO performs the INFO exchange. It is called from within a transaction,
// either on its own or as the verification step at the end of putFile.
func (d *Device) infoIO(ctx context.Context, _ *txn) (r infoResult, err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpINFO)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	// send command:
	err = sendSerialChunked(ctx, d.f, 512, sb)
	if err != nil {
		err = d.FatalError(err)
		return
	}

	// read response:
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		err = d.FatalError(err)
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		err = fmt.Errorf("info: fxpakpro response packet does not contain USBA header")
		err = d.FatalError(err)
		return
	}
	if sb[4] != byte(OpRESPONSE) {
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
	r.rom = string(romB[:clen(romB)])

	versionB := sb[260 : 260+64]
	r.version = string(versionB[:clen(versionB)])

	deviceB := sb[260+64 : 260+64+64]
	r.device = string(deviceB[:clen(deviceB)])

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
