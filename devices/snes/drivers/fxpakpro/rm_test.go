package fxpakpro

import (
	"context"
	"testing"
)

func TestDevice_rm(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "rm /unittest/sub",
			args: args{
				ctx:  context.Background(),
				path: "unittest/sub",
			},
			wantErr: false,
		},
		{
			name: "rm /unittest",
			args: args{
				ctx:  context.Background(),
				path: "unittest",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := d.rm(tt.args.ctx, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("rm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
