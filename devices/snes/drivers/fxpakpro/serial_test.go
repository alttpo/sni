package fxpakpro

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func Test_readExact(t *testing.T) {
	type args struct {
		ctx       context.Context
		f         io.Reader
		chunkSize uint32
		buf       []byte
	}
	tmp := [64]byte{}
	tests := []struct {
		name    string
		args    args
		wantP   uint32
		wantErr bool
	}{
		{
			name: "64 bytes",
			args: args{
				ctx:       context.Background(),
				f:         bytes.NewReader(tmp[:]),
				chunkSize: 64,
				buf:       make([]byte, 64),
			},
			wantP:   64,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP, err := readExactGeneric(tt.args.ctx, tt.args.f, tt.args.chunkSize, tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("readExact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotP != tt.wantP {
				t.Errorf("readExact() gotP = %v, want %v", gotP, tt.wantP)
			}
		})
	}
}
