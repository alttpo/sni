package mock

import (
	"context"
	"github.com/alttpo/snes/timing"
	"sni/devices"
	"sni/protos/sni"
	"sni/util"
	"sync"
	"time"
)

type Device struct {
	lock sync.Mutex

	frameTicker *time.Ticker

	WRAM   []byte
	Memory [0x1000000]byte
}

func (d *Device) Init() {
	// 5,369,317.5/89,341.5 ~= 60.0988 frames / sec ~= 16,639,265.605 ns / frame
	d.frameTicker = time.NewTicker(timing.Frame)

	go func() {
		defer util.Recover()

		for range d.frameTicker.C {
			// increment frame timer:
			d.WRAM[0x1A]++
		}
	}()
}

func (d *Device) IsClosed() bool {
	return false
}

func (d *Device) Close() error {
	panic("implement me")
}

func (d *Device) RequiresMemoryMappingForAddressSpace(ctx context.Context, addressSpace sni.AddressSpace) (bool, error) {
	if addressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	if addressSpace == sni.AddressSpace_SnesABus {
		return false, nil
	}
	return true, nil
}

func (d *Device) RequiresMemoryMappingForAddress(ctx context.Context, address devices.AddressTuple) (bool, error) {
	if address.AddressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	if address.AddressSpace == sni.AddressSpace_SnesABus {
		return false, nil
	}
	return true, nil
}

func (d *Device) MultiReadMemory(context context.Context, reads ...devices.MemoryReadRequest) (mrsps []devices.MemoryReadResponse, err error) {
	// wait 1ms before returning response to simulate the delay of FX Pak Pro device:
	<-time.After(time.Millisecond * 1)

	if len(reads) == 0 {
		return
	}

	mrsps = make([]devices.MemoryReadResponse, 0, len(reads))
	for _, read := range reads {
		data := make([]byte, read.Size)
		copy(data, d.Memory[read.RequestAddress.Address:int(read.RequestAddress.Address)+read.Size])
		mrsps = append(mrsps, devices.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			Data:           data,
		})
	}

	return
}

func (d *Device) MultiWriteMemory(context context.Context, writes ...devices.MemoryWriteRequest) (mrsps []devices.MemoryWriteResponse, err error) {
	// wait 1ms before returning response to simulate the delay of FX Pak Pro device:
	<-time.After(time.Millisecond * 1)

	if len(writes) == 0 {
		return
	}

	mrsps = make([]devices.MemoryWriteResponse, 0, len(writes))
	for _, write := range writes {
		data := write.Data
		dataLen := len(data)

		copy(d.Memory[write.RequestAddress.Address:int(write.RequestAddress.Address)+dataLen], data)

		mrsps = append(mrsps, devices.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			Size:           dataLen,
		})
	}

	return
}

func (d *Device) ResetSystem(ctx context.Context) error {
	panic("implement me")
}

func (d *Device) ResetToMenu(ctx context.Context) error {
	panic("implement me")
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	panic("implement me")
}

func (d *Device) PauseToggle(ctx context.Context) error {
	panic("implement me")
}
