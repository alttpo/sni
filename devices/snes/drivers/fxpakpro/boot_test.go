package fxpakpro

import (
	"context"
	"testing"
)

func TestDevice_boot(t *testing.T) {
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
			name: "boot o2/alttp-jp.smc",
			args: args{
				ctx:  context.Background(),
				path: "o2/alttp-jp.smc",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := d.boot(tt.args.ctx, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("boot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
