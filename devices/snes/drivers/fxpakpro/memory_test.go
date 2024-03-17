package fxpakpro

import (
	"context"
	"github.com/alttpo/snes/asm"
	"log"
	"reflect"
	"sni/devices"
	"sni/protos/sni"
	"testing"
)

func TestGenerateCopyAsm(t *testing.T) {
	type test struct {
		name          string
		args          []devices.MemoryWriteRequest
		wantRemainder []devices.MemoryWriteRequest
	}
	tests := []test{
		{
			name: "Check code_size",
			args: []devices.MemoryWriteRequest{
				{
					RequestAddress: devices.AddressTuple{
						Address:       0xF50010,
						AddressSpace:  sni.AddressSpace_FxPakPro,
						MemoryMapping: sni.MemoryMapping_LoROM,
					},
					Data: []byte{0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				}, {
					RequestAddress: devices.AddressTuple{
						Address:       0xF50010,
						AddressSpace:  sni.AddressSpace_FxPakPro,
						MemoryMapping: sni.MemoryMapping_LoROM,
					},
					Data: []byte{0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F},
				},
			},
		},
	}

	// generate test case:
	{
		d := make([]byte, 512)
		for i := range d {
			d[i] = byte(i)
		}
		tests = append(
			tests,
			test{
				name: "perfect fit",
				args: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF50000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0:0x1e2],
					},
				},
				wantRemainder: []devices.MemoryWriteRequest(nil),
			},
			test{
				name: "split at $1e2",
				args: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF50000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0:512],
					},
				},
				wantRemainder: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF50000 + 0x1e2,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2:],
					},
				},
			},
			test{
				name: "split at 2nd req",
				args: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF50000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0 : 0x1e2-12],
					},
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF51000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2-12:],
					},
				},
				wantRemainder: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF51000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2-12:],
					},
				},
			},
			test{
				name: "split at 2nd req still",
				args: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF50000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0 : 0x1e2-11],
					},
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF51000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2-12:],
					},
				},
				wantRemainder: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF51000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2-12:],
					},
				},
			},
			test{
				name: "split at 1st req",
				args: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF50000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0 : 0x1e2-13],
					},
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF51000,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2-12:],
					},
				},
				wantRemainder: []devices.MemoryWriteRequest{
					{
						RequestAddress: devices.AddressTuple{
							Address:       0xF51000 + 1,
							AddressSpace:  sni.AddressSpace_FxPakPro,
							MemoryMapping: sni.MemoryMapping_LoROM,
						},
						Data: d[0x1e2-12+1:],
					},
				},
			},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := [512]byte{}
			a := asm.NewEmitter(code[:], true)
			remainder := GenerateCopyAsm(a, tt.args...)
			a.WriteTextTo(log.Writer())

			if !reflect.DeepEqual(remainder, tt.wantRemainder) {
				t.Errorf("GenerateCopyAsm remainder %v wantRemainder %v", remainder, tt.wantRemainder)
				return
			}
		})
	}
}

func TestDevice_MultiWriteMemory(t *testing.T) {
	d := openAutoCloseableDevice(t)
	defer d.Close()

	ctx := context.Background()

	var err error
	var rsp []devices.MemoryWriteResponse
	rsp, err = d.MultiWriteMemory(ctx, devices.MemoryWriteRequest{
		RequestAddress: devices.AddressTuple{
			Address:       0xF5FFFE,
			AddressSpace:  sni.AddressSpace_FxPakPro,
			MemoryMapping: sni.MemoryMapping_LoROM,
		},
		Data: []byte{0x55},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = rsp
}
