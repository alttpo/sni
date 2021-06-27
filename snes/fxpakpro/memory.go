package fxpakpro

import (
	"bytes"
	"context"
	"fmt"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/asm"
	"sni/snes/mapping"
)

func (d *Device) MultiReadMemory(
	ctx context.Context,
	reads ...snes.MemoryReadRequest,
) (mrsp []snes.MemoryReadResponse, err error) {
	// make all the response structs and preallocate Data buffers:
	mrsp = make([]snes.MemoryReadResponse, len(reads))
	for j, read := range reads {
		mrsp[j] = snes.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       0,
				AddressSpace:  sni.AddressSpace_FxPakPro,
				MemoryMapping: read.RequestAddress.MemoryMapping,
			},
			Data: make([]byte, read.Size),
		}

		mrsp[j].DeviceAddress.Address, err = mapping.TranslateAddress(
			read.RequestAddress,
			sni.AddressSpace_FxPakPro,
		)
		if err != nil {
			return nil, err
		}
	}

	// Break up larger reads (> 255 bytes) into 255-byte chunks:
	type chunk struct {
		request  int
		vget     [4]byte
		offset   int
		rspStart int
		rspEnd   int
	}
	chunks := make([]chunk, 0, 8)

	sendChunks := func() {
		if len(chunks) > 8 {
			panic(fmt.Errorf("VGET cannot use more than 8 chunks"))
		}

		sb := make([]byte, 64)
		sb[0] = byte('U')
		sb[1] = byte('S')
		sb[2] = byte('B')
		sb[3] = byte('A')
		sb[4] = byte(OpVGET)
		sb[5] = byte(SpaceSNES)
		sb[6] = byte(FlagDATA64B | FlagNORESP)

		total := 0
		sp := sb[32:]
		for _, chunk := range chunks {
			copy(sp, chunk.vget[:])
			sp = sp[4:]
			total += int(chunk.vget[0])
		}

		d.lock.Lock()
		err = sendSerial(d.f, sb)
		if err != nil {
			_ = d.Close()
			d.lock.Unlock()
			return
		}

		// calculate expected number of packets:
		packets := total / 64
		remainder := total & 63
		if remainder > 0 {
			packets++
		}

		// read the expected number of 64-byte packets:
		expected := packets * 64
		rsp := make([]byte, expected)
		err = recvSerial(d.f, rsp, expected)
		if err != nil {
			_ = d.Close()
			d.lock.Unlock()
			return
		}
		d.lock.Unlock()

		// shrink down to exact size:
		rsp = rsp[0:total]
		for _, chunk := range chunks {
			// copy response data:
			copy(mrsp[chunk.request].Data[chunk.offset:], rsp[chunk.rspStart:chunk.rspEnd])
		}
	}

	rspStart := 0
	for j, request := range reads {
		startAddr := mrsp[j].DeviceAddress.Address
		addr := startAddr
		size := request.Size

		for size > 0 {
			chunkSize := 255
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks = append(chunks, chunk{
				request: j,
				vget: [4]byte{
					byte(chunkSize),
					byte((addr >> 16) & 0xFF),
					byte((addr >> 8) & 0xFF),
					byte((addr >> 0) & 0xFF),
				},
				// target offset to write to in Data[] for MemoryReadResponse:
				offset: int(addr - startAddr),
				// source offset to read from in VGET response:
				rspStart: rspStart,
				rspEnd:   rspStart + chunkSize,
			})
			rspStart += chunkSize

			if len(chunks) == 8 {
				sendChunks()
				// reset chunks:
				chunks = chunks[0:0]
				rspStart = 0
			}

			size -= 255
			addr += 255
		}
	}

	if len(chunks) > 0 {
		sendChunks()
	}

	return
}

func (d *Device) MultiWriteMemory(
	ctx context.Context,
	writes ...snes.MemoryWriteRequest,
) (mrsp []snes.MemoryWriteResponse, err error) {
	// make all the response structs:
	mrsp = make([]snes.MemoryWriteResponse, len(writes))
	for j, write := range writes {
		mrsp[j] = snes.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       0,
				AddressSpace:  sni.AddressSpace_FxPakPro,
				MemoryMapping: write.RequestAddress.MemoryMapping,
			},
			Size: len(write.Data),
		}

		mrsp[j].DeviceAddress.Address, err = mapping.TranslateAddress(
			write.RequestAddress,
			sni.AddressSpace_FxPakPro,
		)
		if err != nil {
			return nil, err
		}
	}

	// pick out WRAM writes:
	wramWrites := make([]snes.MemoryWriteRequest, 0, len(writes))

	// Break up larger writes (> 255 bytes) into 255-byte chunks:
	chunks := make([]vputChunk, 0, 8)
	for j, request := range writes {
		startAddr := mrsp[j].DeviceAddress.Address

		// separate out WRAM writes to be handled specially:
		if startAddr >= 0xF50000 && startAddr < 0xF70000 {
			wramWrites = append(wramWrites, snes.MemoryWriteRequest{
				RequestAddress: mrsp[j].DeviceAddress,
				Data:           request.Data,
			})
			continue
		}

		addr := startAddr
		size := len(request.Data)

		for size > 0 {
			chunkSize := 255
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks = append(chunks, vputChunk{
				addr: addr,
				// target offset to write to in Data[] for MemoryWriteResponse:
				data: request.Data[int(addr-startAddr) : int(addr-startAddr)+chunkSize],
			})

			if len(chunks) == 8 {
				err = d.vput(SpaceSNES, chunks)
				if err != nil {
					return
				}
				// reset chunks:
				chunks = chunks[0:0]
			}

			size -= 255
			addr += 255
		}
	}

	if len(chunks) > 0 {
		err = d.vput(SpaceSNES, chunks)
		if err != nil {
			return
		}
	}

	// handle WRAM writes using NMI EXE feature of fxpakpro:
	for _, write := range wramWrites {
		var a asm.Emitter
		a.Code = &bytes.Buffer{}

		// write $00 to $2C00 disables NMI vector:
		a.Code.WriteByte(0)

		// generate a copy routine to write data into WRAM:
		GenerateCopyAsm(&a, write.RequestAddress.Address, write.Data)

		// send PUT command to CMD space:
		err = d.put(0x2C00, SpaceCMD, a.Code.Bytes())
		if err != nil {
			return
		}

		// enable the NMI EXE:
		err = d.vput(SpaceCMD, []vputChunk{{addr: 0x2C00, data: []byte{1}}})
		if err != nil {
			return
		}
	}

	return
}

func GenerateCopyAsm(a *asm.Emitter, targetFXPakProAddress uint32, data []byte) {
	size := uint16(len(data))

	// codeSize represents the total size of ASM code below:
	const codeSize = 0x21

	srcOffset := uint16(0x2C01 + codeSize)
	destOffs := uint16(targetFXPakProAddress & 0xFFFF)
	// FX Pak Pro WRAM addresses are either bank $F5 or $F6:
	destBank := uint8(0x7E + (targetFXPakProAddress-0xF5_0000)>>16)

	a.SetBase(0x002C01)
	a.Comment("preserve registers:")
	a.REP(0x30)
	a.PHA()
	a.PHX()
	a.PHY()

	a.Comment(fmt.Sprintf("transfer $%04x bytes from $00:%04x to $%02x:%04x", size, srcOffset, destBank, destOffs))
	// A - Specifies the amount of bytes to transfer, minus 1
	a.LDA_imm16_w(size - 1)
	// X - Specifies the high and low bytes of the data source memory address
	a.LDX_imm16_w(srcOffset)
	// Y - Specifies the high and low bytes of the destination memory address
	a.LDY_imm16_w(destOffs)
	a.MVN(0x00, destBank)

	a.Comment("disable NMI vector override:")
	a.SEP(0x30)
	a.LDA_imm8_b(0x00)
	a.STA_long(0x002C00)
	a.REP(0x30)

	a.Comment("restore registers:")
	a.PLY()
	a.PLX()
	a.PLA()

	a.Comment("jump to original NMI:")
	a.JMP_indirect(0xFFEA)

	// bug check: make sure emitted code is the expected size
	if actual, expected := a.Code.Len(), codeSize; actual != expected {
		panic(fmt.Errorf("bug check: emitted code size %d != %d", actual, expected))
	}

	// copy in the data to be written to WRAM:
	a.EmitBytes(data)
}
