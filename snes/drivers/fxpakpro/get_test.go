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
