package util

import "testing"

func TestBankToLinear(t *testing.T) {
	type args struct {
		addr uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "Bank 00, Offset 0000",
			args: args{
				addr: 0x000000,
			},
			want: 0x000000,
		},
		{
			name: "Bank 00, Offset 8000",
			args: args{
				addr: 0x008000,
			},
			want: 0x000000,
		},
		{
			name: "Bank 00, Offset 7FFF",
			args: args{
				addr: 0x007FFF,
			},
			want: 0x007FFF,
		},
		{
			name: "Bank 00, Offset FFFF",
			args: args{
				addr: 0x00FFFF,
			},
			want: 0x007FFF,
		},
		{
			name: "Bank 01, Offset 0000",
			args: args{
				addr: 0x010000,
			},
			want: 0x008000,
		},
		{
			name: "Bank 01, Offset 8000",
			args: args{
				addr: 0x018000,
			},
			want: 0x008000,
		},
		{
			name: "Bank 01, Offset 7FFF",
			args: args{
				addr: 0x017FFF,
			},
			want: 0x00FFFF,
		},
		{
			name: "Bank 01, Offset FFFF",
			args: args{
				addr: 0x01FFFF,
			},
			want: 0x00FFFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BankToLinear(tt.args.addr); got != tt.want {
				t.Errorf("BankToLinear() = %v, want %v", got, tt.want)
			}
		})
	}
}
