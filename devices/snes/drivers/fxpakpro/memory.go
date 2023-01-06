package fxpakpro

import (
	"context"
	"fmt"
	"github.com/alttpo/snes/asm"
	"github.com/alttpo/snes/timing"
	"sni/devices"
	"sni/devices/snes/mapping"
	"sni/protos/sni"
	"time"
)

type subspace int

const (
	spaceSNES subspace = 0
	spaceCMD  subspace = 1
)

func (d *Device) RequiresMemoryMappingForAddressSpace(ctx context.Context, addressSpace sni.AddressSpace) (bool, error) {
	if addressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	if addressSpace == sni.AddressSpace_FxPakPro {
		return false, nil
	}
	return true, nil
}

func (d *Device) RequiresMemoryMappingForAddress(ctx context.Context, address devices.AddressTuple) (bool, error) {
	if address.AddressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	if address.AddressSpace == sni.AddressSpace_FxPakPro {
		return false, nil
	}
	return true, nil
}

func (d *Device) MultiReadMemory(
	ctx context.Context,
	reads ...devices.MemoryReadRequest,
) (mrsp []devices.MemoryReadResponse, err error) {
	// VGETs can only be submitted for one Space at a time so keep track of possibly two VGETs if the Spaces are mixed
	// in the `reads` slice:
	chunks := [2][]vgetChunk{
		make([]vgetChunk, 0, 8), // spaceSNES
		make([]vgetChunk, 0, 8), // spaceCMD
	}

	// make all the response structs and preallocate Data buffers:
	mrsp = make([]devices.MemoryReadResponse, len(reads))
	for j, read := range reads {
		mrsp[j] = devices.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			DeviceAddress: devices.AddressTuple{
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

	subctx := ctx
	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
		subctx = context.WithValue(ctx, lockedKey, &struct{}{})
	}

	// Break up larger reads (> 255 bytes) into 255-byte chunks:
	for j, request := range reads {
		startAddr := mrsp[j].DeviceAddress.Address

		// determine the pak Space to read from:
		pakSpace := SpaceSNES
		space := spaceSNES
		if startAddr>>24 == 0x01 {
			pakSpace = SpaceCMD
			space = spaceCMD
			startAddr &= 0x00_FFFFFF
		}

		addr := startAddr
		size := request.Size

		for size > 0 {
			chunkSize := 255
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks[space] = append(chunks[space], vgetChunk{
				target: mrsp[j].Data[int(addr-startAddr):],
				size:   byte(chunkSize),
				addr:   addr,
			})

			if len(chunks[space]) == 8 {
				err = d.vget(subctx, pakSpace, chunks[space]...)
				if err != nil {
					return
				}

				// reset chunks:
				chunks[space] = chunks[space][0:0]
			}

			size -= 255
			addr += 255
		}
	}

	if len(chunks[spaceSNES]) > 0 {
		err = d.vget(subctx, SpaceSNES, chunks[spaceSNES]...)
		if err != nil {
			return
		}
	}

	if len(chunks[spaceCMD]) > 0 {
		err = d.vget(subctx, SpaceCMD, chunks[spaceCMD]...)
		if err != nil {
			return
		}
	}

	return
}

func (d *Device) MultiWriteMemory(
	ctx context.Context,
	writes ...devices.MemoryWriteRequest,
) (mrsp []devices.MemoryWriteResponse, err error) {
	// VPUTs can only be submitted for one Space at a time so keep track of possibly two VPUTs if the Spaces are mixed
	// in the `reads` slice:
	chunks := [2][]vputChunk{
		make([]vputChunk, 0, 8), // spaceSNES
		make([]vputChunk, 0, 8), // spaceCMD
	}

	// make all the response structs:
	mrsp = make([]devices.MemoryWriteResponse, len(writes))
	for j, write := range writes {
		mrsp[j] = devices.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: devices.AddressTuple{
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

	subctx := ctx
	if shouldLock(ctx) {
		// lock the device for this entire sequence to avoid interruptions:
		defer d.lock.Unlock()
		d.lock.Lock()
		subctx = context.WithValue(ctx, lockedKey, &struct{}{})
	}

	// pick out WRAM writes:
	wramWrites := make([]devices.MemoryWriteRequest, 0, len(writes))

	// Break up larger writes (> 255 bytes) into 255-byte chunks:
	for j, request := range writes {
		startAddr := mrsp[j].DeviceAddress.Address

		// separate out WRAM writes to be handled specially:
		if startAddr >= 0xF50000 && startAddr < 0xF70000 {
			wramWrites = append(wramWrites, devices.MemoryWriteRequest{
				RequestAddress: mrsp[j].DeviceAddress,
				Data:           request.Data,
			})
			continue
		}

		pakSpace := SpaceSNES
		space := spaceSNES
		if startAddr>>24 == 0x01 {
			pakSpace = SpaceCMD
			space = spaceCMD
			startAddr &= 0x00_FFFFFF
		}

		addr := startAddr
		size := len(request.Data)

		for size > 0 {
			chunkSize := 255
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks[space] = append(chunks[space], vputChunk{
				addr: addr,
				// target offset to write to in Data[] for MemoryWriteResponse:
				data: request.Data[int(addr-startAddr) : int(addr-startAddr)+chunkSize],
			})

			if len(chunks[space]) == 8 {
				err = d.vput(subctx, pakSpace, chunks[space]...)
				if err != nil {
					return
				}
				// reset chunks:
				chunks[space] = chunks[space][0:0]
			}

			size -= 255
			addr += 255
		}
	}

	if len(chunks[spaceSNES]) > 0 {
		err = d.vput(subctx, SpaceSNES, chunks[spaceSNES]...)
		if err != nil {
			return
		}
	}

	if len(chunks[spaceCMD]) > 0 {
		err = d.vput(subctx, SpaceCMD, chunks[spaceCMD]...)
		if err != nil {
			return
		}
	}

	// handle WRAM writes using NMI EXE feature of fxpakpro:
	if len(wramWrites) > 0 {
		code := [1024]byte{}
		a := asm.NewEmitter(code[:], true)

		// generate a copy routine to write data into WRAM:
		GenerateCopyAsm(a, wramWrites...)

		//a.WriteTextTo(log.Writer())

		if actual, expected := a.Len(), 1024; actual > expected {
			return nil, fmt.Errorf(
				"fxpakpro: too much WRAM data for the snescmd buffer; %d > %d",
				actual,
				expected,
			)
		}

		chunks := make([]vputChunk, 0, 8)
		startAddr := uint32(0x2C00)
		addr := startAddr
		data := code[:a.Len()]
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

		if actual, expected := len(chunks), 8; actual > expected {
			return nil, fmt.Errorf(
				"fxpakpro: too many VPUT chunks to write WRAM data with; %d > %d",
				actual,
				expected,
			)
		}

		// await 5 seconds in game-frames for NMI EXE:
		awaitctx, awaitcancel := context.WithTimeout(subctx, timing.Frame*60*5)
		defer awaitcancel()

		// VGET to await NMI EXE availability:
		{
			var ok bool
			ok, err = d.awaitNMIEXE(awaitctx)
			if err != nil {
				err = fmt.Errorf("fxpakpro: could not acquire NMI EXE pre-write: %w", err)
				return
			}
			if !ok {
				err = fmt.Errorf("fxpakpro: could not acquire NMI EXE pre-write")
				return
			}
		}

		// VPUT command to CMD space:
		err = d.vput(awaitctx, SpaceCMD, chunks...)
		if err != nil {
			err = fmt.Errorf("fxpakpro: could not VPUT to NMI EXE: %w", err)
			return
		}

		// await NMI EXE availability to validate the write was completed:
		{
			var ok bool
			ok, err = d.awaitNMIEXE(awaitctx)
			if err != nil {
				err = fmt.Errorf("fxpakpro: could not acquire NMI EXE post-write: %w", err)
				return
			}
			if !ok {
				err = fmt.Errorf("fxpakpro: could not acquire NMI EXE post-write")
				return
			}
		}
	}

	return
}

func (d *Device) awaitNMIEXE(ctx context.Context) (ok bool, err error) {
	check := make([]byte, 1)

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(time.Second * 5)
	}

	for time.Now().Before(deadline) {
		tmpctx, tmpcancel := context.WithTimeout(ctx, time.Second)
		err = d.vget(tmpctx, SpaceCMD, vgetChunk{addr: 0x2C00, size: 1, target: check})
		tmpcancel()
		if err != nil {
			return
		}
		if check[0] == 0 {
			ok = true
			err = nil
			return
		}
	}

	err = context.DeadlineExceeded
	return
}

func GenerateCopyAsm(a *asm.Emitter, writes ...devices.MemoryWriteRequest) {
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
	if actual, expected := a.Len(), expectedCodeSize; actual != expected {
		panic(fmt.Errorf("bug check: emitted code size %d != %d", actual, expected))
	}

	// copy in the data to be written to WRAM:
	for _, write := range writes {
		a.EmitBytes(write.Data)
	}
}
