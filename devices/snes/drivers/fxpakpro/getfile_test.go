package fxpakpro

import (
	"bytes"
	"context"
	"testing"
)

func DoNotTestDevice_getFile_bug(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	ctx := context.Background()
	{
		n, err := d.putFile(ctx, "romloader/Super Metroid (JU).sfc", 0, bytes.NewReader(nil), nil)
		if err != nil {
			t.Error(err)
			return
		}
		_ = n
	}
	{
		n, err := d.putFile(ctx, "romloader/Super Metroid (JU).sfc", 0, bytes.NewReader(nil), nil)
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

func TestDevice_getFile_oddSizes(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	const size1 = 1023
	const size2 = 1024*17 + 599
	ctx := context.Background()
	{
		n, err := d.putFile(ctx, "unittest/test1.sfc", size1, bytes.NewReader(make([]byte, size1)), nil)
		if err != nil {
			t.Error(err)
			return
		}
		_ = n
	}
	{
		n, err := d.putFile(ctx, "unittest/test2.sfc", size2, bytes.NewReader(make([]byte, size2)), nil)
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
		wantWLen     int
		wantReceived uint32
		wantErr      bool
	}{
		{
			name: "test1.sfc",
			args: args{
				path: "unittest/test1.sfc",
			},
			wantWLen:     size1,
			wantReceived: size1,
			wantErr:      false,
		},
		{
			name: "test2.sfc",
			args: args{
				path: "unittest/test2.sfc",
			},
			wantWLen:     size2,
			wantReceived: size2,
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
			if gotWLen := w.Len(); gotWLen != tt.wantWLen {
				t.Errorf("getFile() gotWLen = %v, want %v", gotWLen, tt.wantWLen)
			}
			if gotReceived != tt.wantReceived {
				t.Errorf("getFile() gotReceived = %v, want %v", gotReceived, tt.wantReceived)
			}
		})
	}
}
