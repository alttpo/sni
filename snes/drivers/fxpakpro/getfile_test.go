package fxpakpro

import (
	"bytes"
	"context"
	"testing"
)

func TestDevice_getFile(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	ctx := context.Background()
	{
		n, err := d.PutFile(ctx, "romloader/Super Metroid (JU).sfc", 0, bytes.NewReader(nil), nil)
		if err != nil {
			t.Error(err)
			return
		}
		_ = n
	}
	{
		n, err := d.PutFile(ctx, "romloader/Super Metroid (JU).sfc", 0, bytes.NewReader(nil), nil)
		if err != nil {
			t.Error(err)
			return
		}
		_ = n
	}

	type args struct {
		path string
	}
	tests := []struct {
		name         string
		args         args
		wantW        string
		wantReceived uint32
		wantErr      bool
	}{
		{
			name: "[0]romloader/Super Metroid (JU).sfc",
			args: args{
				path: "romloader/Super Metroid (JU).sfc",
			},
			wantW:        "",
			wantReceived: 0,
			wantErr:      false,
		},
		{
			name: "[1]romloader/Super Metroid (JU).sfc",
			args: args{
				path: "romloader/Super Metroid (JU).sfc",
			},
			wantW:        "",
			wantReceived: 0,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			gotReceived, err := d.getFile(ctx, tt.args.path, w, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("getFile() gotW = %v, want %v", gotW, tt.wantW)
			}
			if gotReceived != tt.wantReceived {
				t.Errorf("getFile() gotReceived = %v, want %v", gotReceived, tt.wantReceived)
			}
		})
	}
}
