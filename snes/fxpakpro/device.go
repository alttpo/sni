package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"sni/snes"
	"sync"
)

type Device struct {
	lock sync.Mutex
	f    serial.Port

	isClosed bool
}

func (d *Device) IsClosed() bool {
	return d.isClosed
}

func (d *Device) Use(ctx context.Context, user snes.DeviceUser) error {
	if user == nil {
		return nil
	}

	return user(ctx, d)
}

func (d *Device) UseMemory(ctx context.Context, user snes.DeviceMemoryUser) error {
	if user == nil {
		return nil
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	return user(ctx, d)
}

func (d *Device) MultiReadMemory(
	ctx context.Context,
	reads ...snes.MemoryReadRequest,
) (mrsp []snes.MemoryReadResponse, err error) {
	// make all the response structs and preallocate Data buffers:
	mrsp = make([]snes.MemoryReadResponse, len(reads))
	for i, read := range reads {
		mrsp[i].MemoryReadRequest = read
		mrsp[i].Data = make([]byte, read.Size)
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

		err = sendSerial(d.f, sb)
		if err != nil {
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
			return
		}

		// shrink down to exact size:
		rsp = rsp[0:total]
		for _, chunk := range chunks {
			// copy response data:
			copy(mrsp[chunk.request].Data[chunk.offset:], rsp[chunk.rspStart:chunk.rspEnd])
		}
	}

	rspStart := 0
	for reqIndex, request := range reads {
		addr := request.Address
		size := int32(request.Size)

		for size > 0 {
			chunkSize := int32(255)
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks = append(chunks, chunk{
				request: reqIndex,
				vget: [4]byte{
					byte(chunkSize),
					byte((addr >> 16) & 0xFF),
					byte((addr >> 8) & 0xFF),
					byte((addr >> 0) & 0xFF),
				},
				// target offset to write to in Data[] for MemoryReadResponse:
				offset: int(addr - request.Address),
				// source offset to read from in VGET response:
				rspStart: rspStart,
				rspEnd:   rspStart + int(chunkSize),
			})
			rspStart += int(chunkSize)

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
	panic("implement me")
}
