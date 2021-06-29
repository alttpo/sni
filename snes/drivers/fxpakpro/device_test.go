package fxpakpro

import (
	"context"
	"sni/protos/sni"
	"sni/snes"
	"testing"
)

func BenchmarkMemory(b *testing.B) {
	var err error

	// detect fxpakpro devices connected to the system:
	var devices []snes.DeviceDescriptor
	devices, err = driver.Detect()
	if err != nil {
		b.Fatal(err)
	}

	if len(devices) == 0 {
		b.Skip("no fxpakpro devices found to test against")
	}

	uri := &devices[0].Uri

	var d snes.AutoCloseableDevice
	_, d, err = snes.DeviceByUri(uri)
	if err != nil {
		b.Fatal(err)
	}

	// open the device with a single read:
	var rsp []snes.MemoryReadResponse
	rsp, err = d.MultiReadMemory(context.Background(), snes.MemoryReadRequest{
		RequestAddress: snes.AddressTuple{
			Address:       0xF50010,
			AddressSpace:  sni.AddressSpace_FxPakPro,
			MemoryMapping: sni.MemoryMapping_LoROM,
		},
		Size: 1,
	})
	_ = rsp
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.Run("WRAM read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var rsp []snes.MemoryReadResponse
			rsp, err = d.MultiReadMemory(context.Background(), snes.MemoryReadRequest{
				RequestAddress: snes.AddressTuple{
					Address:       0xF50010,
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_LoROM,
				},
				Size: 1,
			})
			_ = rsp
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SRAM read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var rsp []snes.MemoryReadResponse
			rsp, err = d.MultiReadMemory(context.Background(), snes.MemoryReadRequest{
				RequestAddress: snes.AddressTuple{
					Address:       0xE00000,
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_LoROM,
				},
				Size: 1,
			})
			_ = rsp
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SRAM write", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var rsp []snes.MemoryWriteResponse
			rsp, err = d.MultiWriteMemory(context.Background(), snes.MemoryWriteRequest{
				RequestAddress: snes.AddressTuple{
					Address:       0xE07FFF,
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_LoROM,
				},
				Data: []byte{0xFF},
			})
			_ = rsp
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WRAM write", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var rsp []snes.MemoryWriteResponse
			rsp, err = d.MultiWriteMemory(context.Background(), snes.MemoryWriteRequest{
				RequestAddress: snes.AddressTuple{
					Address:       0xF5F340,
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_LoROM,
				},
				Data: []byte{0x04},
			})
			_ = rsp
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
