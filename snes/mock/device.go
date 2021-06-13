package mock

import (
	"context"
	"sni/snes"
	"time"
)

type Device struct {
	frameTicker *time.Ticker

	WRAM   []byte
	Memory [0x1000000]byte
}

func (d *Device) Init() {
	// 5,369,317.5/89,341.5 ~= 60.0988 frames / sec ~= 16,639,265.605 ns / frame
	d.frameTicker = time.NewTicker(16_639_265 * time.Nanosecond)

	go func() {
		for range d.frameTicker.C {
			// increment frame timer:
			d.WRAM[0x1A]++
		}
	}()
}

func (d *Device) IsClosed() bool {
	return false
}

func (d *Device) UseMemory(context context.Context, user snes.MemoryUser) error {
	return user(context, d)
}

func (d *Device) MultiReadMemory(context context.Context, reads ...snes.MemoryReadRequest) (mrsps []snes.MemoryReadResponse, err error) {
	// wait 1ms before returning response to simulate the delay of FX Pak Pro device:
	<-time.After(time.Millisecond * 1)

	if len(reads) == 0 {
		return
	}

	mrsps = make([]snes.MemoryReadResponse, 0, len(reads))
	for _, read := range reads {
		data := make([]byte, read.Size)
		copy(data, d.Memory[read.Address:int(read.Address)+read.Size])
		mrsps = append(mrsps, snes.MemoryReadResponse{
			MemoryReadRequest: read,
			Data:              data,
		})
	}

	return
}

func (d *Device) MultiWriteMemory(context context.Context, writes ...snes.MemoryWriteRequest) (mrsps []snes.MemoryWriteResponse, err error) {
	// wait 1ms before returning response to simulate the delay of FX Pak Pro device:
	<-time.After(time.Millisecond * 1)

	if len(writes) == 0 {
		return
	}

	mrsps = make([]snes.MemoryWriteResponse, 0, len(writes))
	for _, write := range writes {
		data := write.Data
		dataLen := len(data)

		copy(d.Memory[write.Address:int(write.Address)+dataLen], data)

		mrsps = append(mrsps, snes.MemoryWriteResponse{
			Address: write.Address,
			Size:    dataLen,
		})
	}

	return
}
