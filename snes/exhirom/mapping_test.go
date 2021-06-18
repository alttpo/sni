package exhirom

import "testing"

func TestPakAddressToBus(t *testing.T) {
	type args struct {
		pakAddr uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// ROM header shadows:
		{
			name: "ROM header bank $80",
			args: args{
				pakAddr: 0x807FC0,
			},
			want: 0x00FFC0,
		},
		{
			name: "ROM header bank $C0",
			args: args{
				pakAddr: 0xC07FC0,
			},
			want: 0x00FFC0,
		},
		// ROM SlowROM first bank first byte:
		{
			name: "ROM bank $00 first byte",
			args: args{
				pakAddr: 0x800000,
			},
			want: 0x008000,
		},
		{
			name: "ROM bank $1F last byte",
			args: args{
				pakAddr: 0x9FFFFF,
			},
			want: 0x3FFFFF,
		},
		// ROM first page:
		{
			name: "ROM first page bank $00",
			args: args{
				pakAddr: 0x000000,
			},
			want: 0xC00000,
		},
		{
			name: "ROM first page last byte bank $00",
			args: args{
				pakAddr: 0x007FFF,
			},
			want: 0xC07FFF,
		},
		{
			name: "ROM second page last byte bank $00",
			args: args{
				pakAddr: 0x00FFFF,
			},
			want: 0xC0FFFF,
		},
		{
			name: "ROM first page bank $40",
			args: args{
				pakAddr: 0x400000,
			},
			want: 0x400000,
		},
		{
			name: "ROM first page bank $80",
			args: args{
				pakAddr: 0x800000,
			},
			want: 0x008000,
		},
		{
			name: "ROM first page bank $C0",
			args: args{
				// SRAM starts at 0xE00000 in fx pak pro space
				pakAddr: 0xC00000,
			},
			want: 0x008000,
		},
		// ROM last page:
		{
			name: "ROM last page bank $00",
			args: args{
				pakAddr: 0x3FFFFF,
			},
			want: 0xFFFFFF,
		},
		{
			name: "ROM last page bank $40",
			args: args{
				pakAddr: 0x7DFFFF,
			},
			want: 0x7DFFFF,
		},
		{
			name: "ROM last page bank $40",
			args: args{
				pakAddr: 0x7EFFFF,
			},
			want: 0x3FFFFF,
		},
		{
			name: "ROM last page bank $40",
			args: args{
				pakAddr: 0x7FFFFF,
			},
			want: 0x41FFFF,
		},
		{
			name: "ROM last page bank $80",
			args: args{
				pakAddr: 0xBFFFFF,
			},
			want: 0xFFFFFF,
		},
		{
			name: "ROM last page bank $C0",
			args: args{
				// SRAM starts at 0xE00000 in fx pak pro space
				pakAddr: 0xDFFFFF,
			},
			want: 0x3FFFFF,
		},
		// SRAM:
		{
			name: "SRAM $0 bank",
			args: args{
				pakAddr: 0xE00000,
			},
			want: 0xA06000,
		},
		{
			name: "SRAM $0 bank last byte",
			args: args{
				pakAddr: 0xE07FFF,
			},
			want: 0xA37FFF,
		},
		{
			name: "SRAM $1 bank first byte",
			args: args{
				pakAddr: 0xE08000,
			},
			want: 0xA46000,
		},
		{
			name: "SRAM $D bank first byte",
			args: args{
				pakAddr: 0xE68000,
			},
			want: 0xB46000,
		},
		{
			name: "SRAM $D bank last byte",
			args: args{
				pakAddr: 0xE6FFFF,
			},
			want: 0xB77FFF,
		},
		{
			name: "SRAM $E bank",
			args: args{
				pakAddr: 0xE70000,
			},
			want: 0xB86000,
		},
		{
			name: "SRAM $F bank",
			args: args{
				pakAddr: 0xE78000,
			},
			want: 0xBC6000,
		},
		{
			name: "SRAM $F bank last byte",
			args: args{
				pakAddr: 0xE7FFFF,
			},
			want: 0xBF7FFF,
		},
		// mirrored:
		{
			name: "SRAM mirror $0 bank",
			args: args{
				pakAddr: 0xE80000,
			},
			want: 0xA06000,
		},
		{
			name: "SRAM mirror $0 bank last byte",
			args: args{
				pakAddr: 0xE87FFF,
			},
			want: 0xA37FFF,
		},
		{
			name: "SRAM mirror $1 bank first byte",
			args: args{
				pakAddr: 0xE88000,
			},
			want: 0xA46000,
		},
		{
			name: "SRAM mirror $D bank first byte",
			args: args{
				pakAddr: 0xEE8000,
			},
			want: 0xB46000,
		},
		{
			name: "SRAM mirror $D bank last byte",
			args: args{
				pakAddr: 0xEEFFFF,
			},
			want: 0xB77FFF,
		},
		{
			name: "SRAM mirror $E bank",
			args: args{
				pakAddr: 0xEF0000,
			},
			want: 0xB86000,
		},
		{
			name: "SRAM mirror $F bank",
			args: args{
				pakAddr: 0xEF8000,
			},
			want: 0xBC6000,
		},
		{
			name: "SRAM mirror $F bank last byte",
			args: args{
				pakAddr: 0xEFFFFF,
			},
			want: 0xBF7FFF,
		},
		// WRAM:
		{
			name: "WRAM $00000",
			args: args{
				pakAddr: 0xF50000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM $01000",
			args: args{
				pakAddr: 0xF51000,
			},
			want: 0x7E1000,
		},
		{
			name: "WRAM $02000",
			args: args{
				pakAddr: 0xF52000,
			},
			want: 0x7E2000,
		},
		{
			name: "WRAM $0FFFF",
			args: args{
				pakAddr: 0xF5FFFF,
			},
			want: 0x7EFFFF,
		},
		{
			name: "WRAM $10000",
			args: args{
				pakAddr: 0xF60000,
			},
			want: 0x7F0000,
		},
		{
			name: "WRAM $1FFFF",
			args: args{
				pakAddr: 0xF6FFFF,
			},
			want: 0x7FFFFF,
		},
		// WRAM mirrors:
		{
			name: "WRAM mirror 1",
			args: args{
				pakAddr: 0xF70000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 2",
			args: args{
				pakAddr: 0xF90000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 3",
			args: args{
				pakAddr: 0xFB0000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 4",
			args: args{
				pakAddr: 0xFD0000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 5",
			args: args{
				pakAddr: 0xFF0000,
			},
			want: 0x7E0000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PakAddressToBus(tt.args.pakAddr); got != tt.want {
				t.Errorf("PakAddressToBus() = %06x, want %06x", got, tt.want)
			}
		})
	}
}
