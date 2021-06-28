package fxpakpro

import (
	"bytes"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/asm"
	"strings"
	"testing"
)

func TestGenerateCopyAsm(t *testing.T) {
	tests := []struct {
		name string
		args []snes.MemoryWriteRequest
	}{
		{
			name: "Check code_size",
			args: []snes.MemoryWriteRequest{
				{
					RequestAddress: snes.AddressTuple{
						Address:       0xF50010,
						AddressSpace:  sni.AddressSpace_FxPakPro,
						MemoryMapping: sni.MemoryMapping_LoROM,
					},
					Data: []byte{0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				}, {
					RequestAddress: snes.AddressTuple{
						Address:       0xF50010,
						AddressSpace:  sni.AddressSpace_FxPakPro,
						MemoryMapping: sni.MemoryMapping_LoROM,
					},
					Data: []byte{0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a asm.Emitter
			a.Code = &bytes.Buffer{}
			a.Text = &strings.Builder{}
			GenerateCopyAsm(&a, tt.args...)
			t.Log("\n" + a.Text.String())
		})
	}
}
