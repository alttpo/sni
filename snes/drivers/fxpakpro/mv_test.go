package fxpakpro

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

func TestDevice_mv(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	type args struct {
		ctx         context.Context
		path        string
		newFilename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "mv /unittest/test1 test2",
			args: args{
				ctx:         context.Background(),
				path:        "unittest/test1",
				newFilename: "test2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := d.mkdir(tt.args.ctx, filepath.Dir(tt.args.path)); err != nil {
				t.Logf("mkdir() error (ignored) = %v\n", err)
			}

			if _, err := d.putFile(tt.args.ctx, tt.args.path, 1, bytes.NewReader([]byte{1}), nil); err != nil {
				t.Errorf("putFile() error = %v", err)
				return
			}

			if err := d.mv(tt.args.ctx, tt.args.path, tt.args.newFilename); (err != nil) != tt.wantErr {
				t.Errorf("mv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
