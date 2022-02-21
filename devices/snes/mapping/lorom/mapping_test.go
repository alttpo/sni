package lorom

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
			want: 0x80FFC0,
		},
		{
			name: "ROM header bank $40",
			args: args{
				pakAddr: 0x407FC0,
			},
			want: 0x80FFC0,
		},
		{
			name: "ROM header bank $80",
			args: args{
				pakAddr: 0x807FC0,
			},
			want: 0x80FFC0,
		},
		{
			name: "ROM header bank $C0",
			args: args{
				pakAddr: 0xC07FC0,
			},
			want: 0x80FFC0,
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
				pakAddr: 0x7FFFFF,
			},
			want: 0xFFFFFF,
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
			want: 0xBFFFFF,
		},
		// SRAM:
		{
			name: "SRAM $0 bank",
			args: args{
				pakAddr: 0xE00000,
			},
			want: 0xF00000,
		},
		{
			name: "SRAM $0 bank last byte",
			args: args{
				pakAddr: 0xE07FFF,
			},
			want: 0xF07FFF,
		},
		{
			name: "SRAM $1 bank first byte",
			args: args{
				pakAddr: 0xE08000,
			},
			want: 0xF10000,
		},
		{
			name: "SRAM $D bank first byte",
			args: args{
				pakAddr: 0xE68000,
			},
			want: 0xFD0000,
		},
		{
			name: "SRAM $D bank last byte",
			args: args{
				pakAddr: 0xE6FFFF,
			},
			want: 0xFD7FFF,
		},
		{
			name: "SRAM $E bank",
			args: args{
				pakAddr: 0xE70000,
			},
			want: 0xFE0000,
		},
		{
			name: "SRAM $F bank",
			args: args{
				pakAddr: 0xE78000,
			},
			want: 0xFF0000,
		},
		// mirrored:
		{
			name: "SRAM mirror $0 bank",
			args: args{
				pakAddr: 0xE80000,
			},
			want: 0xF00000,
		},
		{
			name: "SRAM mirror $0 bank last byte",
			args: args{
				pakAddr: 0xE87FFF,
			},
			want: 0xF07FFF,
		},
		{
			name: "SRAM mirror $1 bank first byte",
			args: args{
				pakAddr: 0xE88000,
			},
			want: 0xF10000,
		},
		{
			name: "SRAM mirror $D bank first byte",
			args: args{
				pakAddr: 0xEE8000,
			},
			want: 0xFD0000,
		},
		{
			name: "SRAM mirror $D bank last byte",
			args: args{
				pakAddr: 0xEEFFFF,
			},
			want: 0xFD7FFF,
		},
		{
			name: "SRAM mirror $E bank",
			args: args{
				pakAddr: 0xEF0000,
			},
			want: 0xFE0000,
		},
		{
			name: "SRAM mirror $F bank",
			args: args{
				pakAddr: 0xEF8000,
			},
			want: 0xFF0000,
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
				t.Errorf("PakAddressToBus() = %06x, want %06x", got, tt.want)
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
		// ROM header shadows:
		{
			name: "ROM header bank $00",
			args: args{
				busAddr: 0x00FFC0,
			},
			want: 0x007FC0,
		},
		{
			name: "ROM header bank $40",
			args: args{
				busAddr: 0x40FFC0,
			},
			want: 0x007FC0,
		},
		{
			name: "ROM header bank $80",
			args: args{
				busAddr: 0x80FFC0,
			},
			want: 0x007FC0,
		},
		{
			name: "ROM header bank $C0",
			args: args{
				busAddr: 0xC0FFC0,
			},
			want: 0x007FC0,
		},
		// WRAM:
		{
			name: "WRAM $00000",
			args: args{
				busAddr: 0x7E0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM $01000",
			args: args{
				busAddr: 0x7E1000,
			},
			want: 0xF51000,
		},
		{
			name: "WRAM $02000",
			args: args{
				busAddr: 0x7E2000,
			},
			want: 0xF52000,
		},
		{
			name: "WRAM $0FFFF",
			args: args{
				busAddr: 0x7EFFFF,
			},
			want: 0xF5FFFF,
		},
		{
			name: "WRAM $10000",
			args: args{
				busAddr: 0x7F0000,
			},
			want: 0xF60000,
		},
		{
			name: "WRAM $1FFFF",
			args: args{
				busAddr: 0x7FFFFF,
			},
			want: 0xF6FFFF,
		},
		// banks F0-FF:
		{
			name: "ROM $F0 bank",
			args: args{
				busAddr: 0xF08000,
			},
			want: 0x180000,
		},
		{
			name: "ROM $F0 bank last byte",
			args: args{
				busAddr: 0xF0FFFF,
			},
			want: 0x187FFF,
		},
		{
			name: "ROM $FF bank",
			args: args{
				busAddr: 0xFF8000,
			},
			want: 0x1F8000,
		},
		{
			name: "ROM $FF bank last byte",
			args: args{
				busAddr: 0xFFFFFF,
			},
			want: 0x1FFFFF,
		},
		// SRAM:
		{
			name: "SRAM $F0 bank",
			args: args{
				busAddr: 0xF00000,
			},
			want: 0xE00000,
		},
		{
			name: "SRAM $F0 bank last byte",
			args: args{
				busAddr: 0xF07FFF,
			},
			want: 0xE07FFF,
		},
		{
			name: "SRAM $F1 bank first byte",
			args: args{
				busAddr: 0xF10000,
			},
			want: 0xE08000,
		},
		{
			name: "SRAM $FD bank first byte",
			args: args{
				busAddr: 0xFD0000,
			},
			want: 0xE68000,
		},
		{
			name: "SRAM $FD bank last byte",
			args: args{
				busAddr: 0xFD7FFF,
			},
			want: 0xE6FFFF,
		},
		{
			name: "SRAM $FE bank",
			args: args{
				busAddr: 0xFE0000,
			},
			want: 0xE70000,
		},
		{
			name: "SRAM $FF bank",
			args: args{
				busAddr: 0xFF0000,
			},
			want: 0xE78000,
		},
		// WRAM:
		{
			name: "WRAM $80 bank $0000",
			args: args{
				busAddr: 0x800000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM $80 bank $1000",
			args: args{
				busAddr: 0x801000,
			},
			want: 0xF51000,
		},
		// banks 70-7D:
		{
			name: "ROM $70 bank",
			args: args{
				busAddr: 0x708000,
			},
			want: 0x180000,
		},
		{
			name: "ROM $70 bank last byte",
			args: args{
				busAddr: 0x70FFFF,
			},
			want: 0x187FFF,
		},
		{
			name: "ROM $7D bank",
			args: args{
				busAddr: 0x7D8000,
			},
			want: 0x1E8000,
		},
		{
			name: "ROM $7D bank last byte",
			args: args{
				busAddr: 0x7DFFFF,
			},
			want: 0x1EFFFF,
		},
		{
			name: "SRAM $70 bank",
			args: args{
				busAddr: 0x700000,
			},
			want: 0xE00000,
		},
		{
			name: "SRAM $70 bank last byte",
			args: args{
				busAddr: 0x707FFF,
			},
			want: 0xE07FFF,
		},
		{
			name: "SRAM $71 bank first byte",
			args: args{
				busAddr: 0x710000,
			},
			want: 0xE08000,
		},
		{
			name: "SRAM $7D bank first byte",
			args: args{
				busAddr: 0x7D0000,
			},
			want: 0xE68000,
		},
		{
			name: "SRAM $7D bank last byte",
			args: args{
				busAddr: 0x7D7FFF,
			},
			want: 0xE6FFFF,
		},
		// WRAM:
		{
			name: "WRAM $00 bank $0000",
			args: args{
				busAddr: 0x000000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM $00 bank $1000",
			args: args{
				busAddr: 0x001000,
			},
			want: 0xF51000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := BusAddressToPak(tt.args.busAddr); got != tt.want {
				t.Errorf("BusAddressToPak() = %v, want %v", got, tt.want)
			}
		})
	}
}
