package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
	"sni/protos/sni"
	"sni/snes"
)

func (d *Device) listFiles(ctx context.Context, path string) (files []snes.DirEntry, err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpLS)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)

	n := copy(sb[256:], path)
	binary.BigEndian.PutUint32(sb[252:], uint32(n))

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	// send the data to the USB port:
	err = sendSerial(d.f, 512, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	// await the first response packet for error status:
	err = recvSerial(ctx, d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}

	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		return nil, fmt.Errorf("mkdir: fxpakpro response packet does not contain USBA header")
	}

	// fxpakpro `ls` command always returns 1 for size:
	if size := binary.BigEndian.Uint32(sb[252:256]); size != 1 {
		_ = d.Close()
		return nil, fmt.Errorf("ls: fxpakpro response size actual %d, expected 1", size)
	}
	if sb[4] != byte(OpRESPONSE) {
		_ = d.Close()
		return nil, fmt.Errorf("ls: wrong opcode in response packet; got $%02x", sb[4])
	}
	if ec := sb[5]; ec != 0 {
		return nil, fmt.Errorf("ls: %w", fxpakproError(ec))
	}

	files = make([]snes.DirEntry, 0, 10)

recvLoop:
	for {
		iterCtx, iterCancel := context.WithTimeout(ctx, safeTimeout)
		err = recvSerial(iterCtx, d.f, sb, 512)
		iterCancel()
		if err != nil {
			_ = d.Close()
			return
		}

		for i := 0; i < 512; {
			// FF means no more data expected:
			if sb[i] == 0xFF {
				break recvLoop
			}
			// 2 means more data expected in the next packet:
			if sb[i] == 2 {
				continue recvLoop
			}

			file := snes.DirEntry{
				Name: "",
				Type: 0,
			}

			// 0 for directory, 1 for file
			if sb[i] == 0 {
				file.Type = sni.DirEntryType_Directory
			} else if sb[i] == 1 {
				file.Type = sni.DirEntryType_File
			}
			i++

			// read filename with 0-terminator:
			start := i
			for i < 512 && sb[i] != 0 {
				i++
			}
			if i >= 512 {
				return nil, fmt.Errorf("ls: invalid response packet format")
			}
			file.Name = string(sb[start:i])
			i++

			// file size does not come in this response
			files = append(files, file)
		}
	}

	// TODO: go back and fetch file sizes
	// NOTE: there is no way in the protocol to simply check file size. GET requires downloading the entire file.
	//for i, file := range files {
	//	size, err = d.getFile(file.Name)
	//}

	return
}
