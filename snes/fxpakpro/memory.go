package fxpakpro

import (
	"bytes"
	"context"
	"fmt"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/asm"
	"sni/snes/mapping"
	"strings"
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
	chunks := make([]vgetChunk, 0, 8)
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
			chunks = append(chunks, vgetChunk{
				target: mrsp[j].Data[int(addr - startAddr):],
				size: byte(chunkSize),
				addr: addr,
				// source offset to read from in VGET response:
				rspStart: rspStart,
				rspEnd:   rspStart + chunkSize,
			})
			rspStart += chunkSize

			if len(chunks) == 8 {
				err = d.vget(SpaceSNES, chunks)
				if err != nil {
					return
				}

				// reset chunks:
				chunks = chunks[0:0]
				rspStart = 0
			}

			size -= 255
			addr += 255
		}
	}

	if len(chunks) > 0 {
		err = d.vget(SpaceSNES, chunks)
		if err != nil {
			return
		}
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
	if len(wramWrites) > 0 {
		var a asm.Emitter
		a.Code = &bytes.Buffer{}
		a.Text = &strings.Builder{}

		// generate a copy routine to write data into WRAM:
		GenerateCopyAsm(&a, wramWrites...)

		//log.Print("\n" + a.Text.String())

		if actual, expected := a.Code.Len(), 1024; actual > expected {
			return nil, fmt.Errorf(
				"fxpakpro: too much WRAM data for the snescmd buffer; %d > %d",
				actual,
				expected,
			)
		}

		chunks := make([]vputChunk, 0, 8)
		startAddr := uint32(0x2C00)
		addr := startAddr
		data := a.Code.Bytes()
		size := len(data)
		for size > 0 {
			chunkSize := 255
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks = append(chunks, vputChunk{
				addr: addr,
				// target offset to write to in Data[] for MemoryWriteResponse:
				data: data[int(addr-startAddr) : int(addr-startAddr)+chunkSize],
			})

			size -= 255
			addr += 255
		}

		// enable the NMI EXE feature by writing 1:
		//chunks = append(chunks, vputChunk{0x2C00, []byte{1}})

		if actual, expected := len(chunks), 8; actual > expected {
			return nil, fmt.Errorf(
				"fxpakpro: too many VPUT chunks to write WRAM data with; %d > %d",
				actual,
				expected,
			)
		}

		// VPUT command to CMD space:
		err = d.vput(SpaceCMD, chunks)
		if err != nil {
			return
		}
	}

	return
}

func GenerateCopyAsm(a *asm.Emitter, writes ...snes.MemoryWriteRequest) {
	// codeSize represents the total size of ASM code below without MVN blocks:
	const codeSize = 0x1B

	a.SetBase(0x002C00)

	// this NOP slide is necessary to avoid the problematic $2C00 address itself.
	a.NOP()
	a.NOP()

	a.Comment("preserve registers:")

	a.REP(0x30)
	a.PHA()
	a.PHX()
	a.PHY()
	a.PHD()

	// MVN affects B register:
	a.PHB()
	expectedCodeSize := codeSize + (12 * len(writes))
	srcOffs := uint16(0x2C00 + expectedCodeSize)
	for _, write := range writes {
		data := write.Data
		size := uint16(len(data))
		targetFXPakProAddress := write.RequestAddress.Address
		destBank := uint8(0x7E + (targetFXPakProAddress-0xF5_0000)>>16)
		destOffs := uint16(targetFXPakProAddress & 0xFFFF)

		a.Comment(fmt.Sprintf("transfer $%04x bytes from $00:%04x to $%02x:%04x", size, srcOffs, destBank, destOffs))
		// A - Specifies the amount of bytes to transfer, minus 1
		a.LDA_imm16_w(size - 1)
		// X - Specifies the high and low bytes of the data source memory address
		a.LDX_imm16_w(srcOffs)
		// Y - Specifies the high and low bytes of the destination memory address
		a.LDY_imm16_w(destOffs)
		a.MVN(destBank, 0x00)

		srcOffs += size
	}
	a.PLB()

	a.Comment("disable NMI vector override:")
	a.SEP(0x30)
	a.LDA_imm8_b(0x00)
	a.STA_long(0x002C00)

	a.Comment("restore registers:")
	a.REP(0x30)
	a.PLD()
	a.PLY()
	a.PLX()
	a.PLA()

	a.Comment("jump to original NMI:")
	a.JMP_indirect(0xFFEA)

	// bug check: make sure emitted code is the expected size
	if actual, expected := a.Code.Len(), expectedCodeSize; actual != expected {
		panic(fmt.Errorf("bug check: emitted code size %d != %d", actual, expected))
	}

	// copy in the data to be written to WRAM:
	for _, write := range writes {
		a.EmitBytes(write.Data)
	}
}
