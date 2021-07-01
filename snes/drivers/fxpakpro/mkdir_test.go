package fxpakpro

import (
	"context"
	"testing"
)

func TestDevice_mkdir(t *testing.T) {
	d := openExactDevice(t)

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
			name: "mkdir /unittest",
			args: args{
				ctx:  context.Background(),
				path: "unittest",
			},
			wantErr: false,
		},
		{
			name: "mkdir /unittest/sub",
			args: args{
				ctx:  context.Background(),
				path: "unittest/sub",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := d.mkdir(tt.args.ctx, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("mkdir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
