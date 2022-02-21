package fxpakpro

import (
	"context"
	"io"
	"testing"
)

type patternReader struct {
	size uint32
	offs uint32
}

func (r *patternReader) Read(d []byte) (n int, err error) {
	n = 0
	for i := range d {
		if r.offs >= r.size {
			err = io.EOF
			return
		}
		d[i] = byte(r.offs)

		r.offs++
		n++
		if n >= 63 {
			return
		}
	}

	return
}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &patternReader{size: tt.args.size}
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
