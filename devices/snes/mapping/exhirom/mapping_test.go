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
			name: "ROM header bank $00",
			args: args{
				pakAddr: 0x007FC0,
			},
			want: 0xC07FC0,
		},
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
			if got, _ := PakAddressToBus(tt.args.pakAddr); got != tt.want {
				t.Errorf("PakAddressToBus() = 0x%06x, want 0x%06x", got, tt.want)
			}
		})
	}
}

func TestBusAddressToPak(t *testing.T) {
	type args struct {
		busAddr uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// banks $80-FF:
		{
			name: "ROM bank $FE:0000",
			args: args{
				busAddr: 0xFE0000,
			},
			want: 0x3E0000,
		},
		{
			name: "ROM bank $FE:FFFF",
			args: args{
				busAddr: 0xFEFFFF,
			},
			want: 0x3EFFFF,
		},
		{
			name: "ROM bank $FF:0000",
			args: args{
				busAddr: 0xFF0000,
			},
			want: 0x3F0000,
		},
		{
			name: "ROM bank $FF:FFFF",
			args: args{
				busAddr: 0xFFFFFF,
			},
			want: 0x3FFFFF,
		},
		{
			name: "ROM bank $C0:0000",
			args: args{
				busAddr: 0xC00000,
			},
			want: 0x000000,
		},
		{
			name: "ROM bank $C0:FFFF",
			args: args{
				busAddr: 0xC0FFFF,
			},
			want: 0x00FFFF,
		},
		{
			name: "ROM bank $FD:0000",
			args: args{
				busAddr: 0xFD0000,
			},
			want: 0x3D0000,
		},
		{
			name: "ROM bank $FD:FFFF",
			args: args{
				busAddr: 0xFDFFFF,
			},
			want: 0x3DFFFF,
		},
		{
			name: "ROM bank $A0:8000",
			args: args{
				busAddr: 0xA08000,
			},
			want: 0x100000,
		},
		{
			name: "ROM bank $A1:8000",
			args: args{
				busAddr: 0xA18000,
			},
			want: 0x108000,
		},
		{
			name: "SRAM bank $A0:6000",
			args: args{
				busAddr: 0xA06000,
			},
			want: 0xE00000,
		},
		{
			name: "SRAM bank $A1:6000",
			args: args{
				busAddr: 0xA16000,
			},
			want: 0xE02000,
		},
		{
			name: "SRAM bank $BF:6000",
			args: args{
				busAddr: 0xBF6000,
			},
			want: 0xE3E000,
		},
		{
			name: "SRAM bank $BF:7FFF",
			args: args{
				busAddr: 0xBF7FFF,
			},
			want: 0xE3FFFF,
		},
		{
			name: "ROM bank $80:8000",
			args: args{
				busAddr: 0x808000,
			},
			want: 0x000000,
		},
		{
			name: "ROM bank $81:8000",
			args: args{
				busAddr: 0x818000,
			},
			want: 0x008000,
		},
		{
			name: "WRAM bank $80:0000",
			args: args{
				busAddr: 0x800000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $80:1FFF",
			args: args{
				busAddr: 0x801FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $9F:0000",
			args: args{
				busAddr: 0x9F0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $9F:1FFF",
			args: args{
				busAddr: 0x9F1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $A0:0000",
			args: args{
				busAddr: 0xA00000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $A0:1FFF",
			args: args{
				busAddr: 0xA01FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $BF:0000",
			args: args{
				busAddr: 0xBF0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $BF:1FFF",
			args: args{
				busAddr: 0xBF1FFF,
			},
			want: 0xF51FFF,
		},
		// WRAM banks 7E-7F:
		{
			name: "WRAM bank $7E:0000",
			args: args{
				busAddr: 0x7E0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $7E:1FFF",
			args: args{
				busAddr: 0x7E1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $7E:2000",
			args: args{
				busAddr: 0x7E2000,
			},
			want: 0xF52000,
		},
		{
			name: "WRAM bank $7E:3FFF",
			args: args{
				busAddr: 0x7E3FFF,
			},
			want: 0xF53FFF,
		},
		{
			name: "WRAM bank $7E:FFFF",
			args: args{
				busAddr: 0x7EFFFF,
			},
			want: 0xF5FFFF,
		},
		{
			name: "WRAM bank $7F:0000",
			args: args{
				busAddr: 0x7F0000,
			},
			want: 0xF60000,
		},
		{
			name: "WRAM bank $7F:FFFF",
			args: args{
				busAddr: 0x7FFFFF,
			},
			want: 0xF6FFFF,
		},
		// banks 00-7D:
		{
			name: "ROM bank $40:0000",
			args: args{
				busAddr: 0x400000,
			},
			want: 0x400000,
		},
		{
			name: "ROM bank $40:FFFF",
			args: args{
				busAddr: 0x40FFFF,
			},
			want: 0x40FFFF,
		},
		{
			name: "ROM bank $7D:0000",
			args: args{
				busAddr: 0x7D0000,
			},
			want: 0x7D0000,
		},
		{
			name: "ROM bank $7D:FFFF",
			args: args{
				busAddr: 0x7DFFFF,
			},
			want: 0x7DFFFF,
		},
		// banks 3E-3F:
		{
			name: "ROM bank $3E:8000",
			args: args{
				busAddr: 0x3E8000,
			},
			want: 0x5F0000,
		},
		{
			name: "ROM bank $3F:8000",
			args: args{
				busAddr: 0x3F8000,
			},
			want: 0x5F8000,
		},
		// banks 20-3D:
		{
			name: "ROM bank $20:8000",
			args: args{
				busAddr: 0x208000,
			},
			want: 0x500000,
		},
		{
			name: "ROM bank $21:8000",
			args: args{
				busAddr: 0x218000,
			},
			want: 0x508000,
		},
		// banks 00-1F:
		{
			name: "ROM bank $00:FFC0",
			args: args{
				busAddr: 0x00FFC0,
			},
			want: 0x407FC0,
		},
		{
			name: "ROM bank $00:8000",
			args: args{
				busAddr: 0x008000,
			},
			want: 0x400000,
		},
		{
			name: "ROM bank $00:FFFF",
			args: args{
				busAddr: 0x00FFFF,
			},
			want: 0x407FFF,
		},
		{
			name: "ROM bank $1F:8000",
			args: args{
				busAddr: 0x1F8000,
			},
			want: 0x4F8000,
		},
		{
			name: "ROM bank $1F:FFFF",
			args: args{
				busAddr: 0x1FFFFF,
			},
			want: 0x4FFFFF,
		},
		// WRAM:
		{
			name: "WRAM bank $00:0000",
			args: args{
				busAddr: 0x000000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $00:1FFF",
			args: args{
				busAddr: 0x001FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $1F:0000",
			args: args{
				busAddr: 0x1F0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $1F:1FFF",
			args: args{
				busAddr: 0x1F1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $20:0000",
			args: args{
				busAddr: 0x200000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $20:1FFF",
			args: args{
				busAddr: 0x201FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $3F:0000",
			args: args{
				busAddr: 0x3F0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $3F:1FFF",
			args: args{
				busAddr: 0x3F1FFF,
			},
			want: 0xF51FFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := BusAddressToPak(tt.args.busAddr); got != tt.want {
				t.Errorf("BusAddressToPak() = 0x%06x, want 0x%06x", got, tt.want)
			}
		})
	}
}
