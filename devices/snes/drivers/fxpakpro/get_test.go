package fxpakpro

import (
	"context"
	"testing"
)

func TestDevice_get(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	type args struct {
		ctx     context.Context
		space   space
		address uint32
		size    uint32
	}
	tests := []struct {
		name        string
		args        args
		wantDataLen int
		wantErr     bool
	}{
		{
			name: "",
			args: args{
				space:   SpaceSNES,
				address: 0xF50000,
				size:    0x200,
			},
			wantDataLen: 0x200,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gotData, err := d.get(ctx, tt.args.space, tt.args.address, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotData) != tt.wantDataLen {
				t.Errorf("get() len(gotData) = %v, want %v", len(gotData), tt.wantDataLen)
			}
		})
	}
}

func BenchmarkDevice_get(b *testing.B) {
	d := openExactDevice(b)
	defer d.Close()

	b.Run("GET", func(b *testing.B) {
		const byteCount = 0x2000
		b.SetBytes(int64(byteCount + 0x200))
		ctx := context.Background()
		for n := 0; n < b.N; n++ {
			_, err := d.get(ctx, SpaceSNES, 0xF50000, uint32(byteCount))
			if err != nil {
				b.Error(err)
				return
			}
		}
	})
}
