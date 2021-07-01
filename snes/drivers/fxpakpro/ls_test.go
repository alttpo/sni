package fxpakpro

import (
	"reflect"
	"sni/protos/sni"
	"sni/snes"
	"testing"
)

func TestDevice_listFiles(t *testing.T) {
	d := openExactDevice(t)

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
			wantFiles: []snes.DirEntry{
				{Name: "o2", Type: sni.DirEntryType_Directory, Size: 0},
				{Name: "poop", Type: sni.DirEntryType_Directory, Size: 0},
				{Name: "roms", Type: sni.DirEntryType_Directory, Size: 0},
				{Name: "System Volume Information", Type: sni.DirEntryType_Directory, Size: 0},
				{Name: "sd2snes", Type: sni.DirEntryType_Directory, Size: 0},
			},
			wantErr: false,
		},
		{
			name: "list /o2",
			args: args{
				path: "/o2",
			},
			wantFiles: []snes.DirEntry{
				{Name: ".", Type: sni.DirEntryType_Directory, Size: 0},
				{Name: "..", Type: sni.DirEntryType_Directory, Size: 0},
				{Name: "lttp.smc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "lttphack-vanilla.sfc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "patched.smc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "lttpj.smc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "lttphack-v13.1.1-emulator-vanillahud.sfc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "alttp-jp.smc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "lttphack-vanilla.smc", Type: sni.DirEntryType_File, Size: 0},
				{Name: "alttpr - noglitches-standard-ganon_gvmrxlwjy5.sfc", Type: sni.DirEntryType_File, Size: 0},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := d.listFiles(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("listFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("listFiles() gotFiles = %#v, want %#v", gotFiles, tt.wantFiles)
			}
		})
	}
}
