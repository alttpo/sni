package fxpakpro

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"sync"
)

type Device struct {
	lock sync.Mutex
	f    serial.Port

	isClosed bool
}

func (d *Device) Init() error {
	return nil
}

func (d *Device) IsClosed() bool {
	return d.isClosed
}

func (d *Device) UseMemory(ctx context.Context, requiredCapabilities []sni.DeviceCapability, user snes.DeviceMemoryUser) error {
	if user == nil {
		return nil
	}

	if ok, err := driver.HasCapabilities(requiredCapabilities...); !ok {
		return err
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	return user(ctx, d)
}

func (d *Device) UseControl(ctx context.Context, requiredCapabilities []sni.DeviceCapability, user snes.DeviceControlUser) error {
	if user == nil {
		return nil
	}

	if ok, err := driver.HasCapabilities(requiredCapabilities...); !ok {
		return err
	}

	defer d.lock.Unlock()
	d.lock.Lock()

	return user(ctx, d)
}

func (d *Device) MultiReadMemory(
	ctx context.Context,
	reads ...snes.MemoryReadRequest,
) (mrsp []snes.MemoryReadResponse, err error) {
	defer func() {
		if err != nil {
			mrsp = nil
			_ = d.f.Close()
			d.isClosed = true
		}
	}()

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
	defer func() {
		if err != nil {
			mrsp = nil
			_ = d.f.Close()
			d.isClosed = true
		}
	}()

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

	// Break up larger writes (> 255 bytes) into 255-byte chunks:
	type chunk struct {
		request int
		vput    [4]byte
		data    []byte
	}
	chunks := make([]chunk, 0, 8)

	sendChunks := func() {
		if len(chunks) > 8 {
			panic(fmt.Errorf("VPUT cannot use more than 8 chunks"))
		}

		sb := make([]byte, 64)
		sb[0] = byte('U')
		sb[1] = byte('S')
		sb[2] = byte('B')
		sb[3] = byte('A')
		sb[4] = byte(OpVPUT)
		sb[5] = byte(SpaceSNES)
		sb[6] = byte(FlagDATA64B | FlagNORESP)

		total := 0
		sp := sb[32:]
		for _, chunk := range chunks {
			copy(sp, chunk.vput[:])
			sp = sp[4:]
			total += int(chunk.vput[0])
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

		// concatenate all accompanying data together in one large slice:
		expected := packets * 64
		whole := make([]byte, expected)
		o := 0
		for _, chunk := range chunks {
			copy(whole[o:], chunk.data)
			o += len(chunk.data)
		}

		// send the expected number of 64-byte packets:
		err = sendSerial(d.f, whole)
		if err != nil {
			return
		}
	}

	for j, request := range writes {
		startAddr := mrsp[j].DeviceAddress.Address
		addr := startAddr
		size := len(request.Data)

		for size > 0 {
			chunkSize := 255
			if size < chunkSize {
				chunkSize = size
			}

			// 4-byte struct: 1 byte size, 3 byte address
			chunks = append(chunks, chunk{
				request: j,
				vput: [4]byte{
					byte(chunkSize),
					byte((addr >> 16) & 0xFF),
					byte((addr >> 8) & 0xFF),
					byte((addr >> 0) & 0xFF),
				},
				// target offset to write to in Data[] for MemoryWriteResponse:
				data: request.Data[int(addr-startAddr) : int(addr-startAddr)+chunkSize],
			})

			if len(chunks) == 8 {
				sendChunks()
				// reset chunks:
				chunks = chunks[0:0]
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

func (d *Device) ResetSystem(ctx context.Context) error {
	panic("implement me")
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	panic("implement me")
}

func (d *Device) PauseToggle(ctx context.Context) error {
	panic("implement me")
}
