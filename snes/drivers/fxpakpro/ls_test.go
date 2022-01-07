package fxpakpro

import (
	"context"
	"sni/snes"
	"testing"
)

func TestDevice_listFiles(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	type args struct {
		path string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []snes.DirEntry
		wantErr   bool
	}{
		{
			name: "list /",
			args: args{
				path: "/",
			},
			wantErr: false,
		},
		{
			name: "list ''",
			args: args{
				path: "",
			},
			wantErr: false,
		},
		{
			name: "list 'o2'",
			args: args{
				path: "o2",
			},
			wantErr: false,
		},
		{
			name: "list 'romloader'",
			args: args{
				path: "romloader",
			},
			wantErr: false,
		},
		{
			name: "list 'o2'",
			args: args{
				path: "o2",
			},
			wantErr: false,
		},
		{
			name: "list ''",
			args: args{
				path: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := d.listFiles(context.Background(), tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("listFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(gotFiles)
		})
	}
}
