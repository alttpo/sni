package fxpakpro

import (
	"bytes"
	"sni/snes/asm"
	"strings"
	"testing"
)

func TestGenerateCopyAsm(t *testing.T) {
	type args struct {
		targetFXPakProAddress uint32
		data                  []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Check code_size",
			args: args{
				targetFXPakProAddress: 0xF50010,
				data:                  []byte{0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a asm.Emitter
			a.Code = &bytes.Buffer{}
			a.Text = &strings.Builder{}
			a.SetBase(0x002C01)
			GenerateCopyAsm(&a, tt.args.targetFXPakProAddress, tt.args.data)
			t.Log("\n" + a.Text.String())
		})
	}
}
