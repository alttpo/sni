package fxpakpro

import (
	"bytes"
	"context"
	"testing"
)

func TestDevice_putFile(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	ctx := context.Background()
	type args struct {
		path string
		size uint32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "511.sfc",
			args: args{
				path: "unittest/test1.sfc",
				size: 511,
			},
			wantErr: false,
		},
		{
			name: "513.sfc",
			args: args{
				path: "unittest/test2.sfc",
				size: 513,
			},
			wantErr: false,
		},
	}

	testdata := [513]byte{}
	for i := range testdata {
		testdata[i] = byte(i & 255)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(testdata[:tt.args.size])
			n, err := d.putFile(ctx, tt.args.path, tt.args.size, r, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("putFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = n

			// make sure a second command works immediately after transfer:
			_, _, _, err = d.info(ctx)
			if err != nil {
				t.Errorf("info() after putFile() failed: %v", err)
			}
		})
	}
}
