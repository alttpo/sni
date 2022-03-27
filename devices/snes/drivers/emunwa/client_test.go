package emunwa

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func Test_parseResponse(t *testing.T) {
	type args struct {
		d []byte
	}
	tests := []struct {
		name      string
		args      args
		wantBin   []byte
		wantAscii []map[string]string
		wantErr   bool
	}{
		{
			name:      "empty success",
			args:      args{[]byte("\n\n")},
			wantBin:   nil,
			wantAscii: []map[string]string{},
			wantErr:   false,
		},
		{
			name:    "ok response",
			args:    args{[]byte("\nok\n")},
			wantBin: nil,
			wantAscii: []map[string]string{
				{"ok": ""},
			},
			wantErr: false,
		},
		{
			name:    "1 item",
			args:    args{[]byte("\na:1\nb:2\n")},
			wantBin: nil,
			wantAscii: []map[string]string{
				{"a": "1", "b": "2"},
			},
			wantErr: false,
		},
		{
			name:    "2 items",
			args:    args{[]byte("\na:1\nb:2\na:4\nb:5\n")},
			wantBin: nil,
			wantAscii: []map[string]string{
				{"a": "1", "b": "2"},
				{"a": "4", "b": "5"},
			},
			wantErr: false,
		},
		{
			name:      "1 byte",
			args:      args{[]byte("\x00\x00\x00\x00\x01A")},
			wantBin:   []byte{'A'},
			wantAscii: nil,
			wantErr:   false,
		},
		{
			name:      "2 bytes 1 ignored",
			args:      args{[]byte("\x00\x00\x00\x00\x01AB")},
			wantBin:   []byte{'A'},
			wantAscii: nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.Write(tt.args.d)
			r := bufio.NewReader(&buf)

			c := Client{}
			gotBin, gotAscii, err := c.parseResponse(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBin, tt.wantBin) {
				t.Errorf("parseResponse() gotBin = %v, want %v", gotBin, tt.wantBin)
			}
			if !reflect.DeepEqual(gotAscii, tt.wantAscii) {
				t.Errorf("parseResponse() gotAscii = %v, want %v", gotAscii, tt.wantAscii)
			}
		})
	}
}
