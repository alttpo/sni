package util

import (
	"strings"
	"testing"
)

func TestIndenter_Write(t *testing.T) {
	type fields struct {
		indent []byte
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr bool
		wantStr string
	}{
		{
			name: "single line indent",
			fields: fields{
				indent: []byte("  "),
			},
			args: args{
				p: []byte("abc\ndef"),
			},
			wantN:   7,
			wantErr: false,
			wantStr: "  abc\n  def",
		},
		{
			name: "retain empty indent",
			fields: fields{
				indent: []byte("  "),
			},
			args: args{
				p: []byte("\n\ndef"),
			},
			wantN:   5,
			wantErr: false,
			wantStr: "  \n  \n  def",
		},
		{
			name: "no trailing empty indent",
			fields: fields{
				indent: []byte("  "),
			},
			args: args{
				p: []byte("def\n"),
			},
			wantN:   4,
			wantErr: false,
			wantStr: "  def\n",
		},
		{
			name: "leading empty indent",
			fields: fields{
				indent: []byte("  "),
			},
			args: args{
				p: []byte("\ndef\n"),
			},
			wantN:   5,
			wantErr: false,
			wantStr: "  \n  def\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			w := NewIndenter(sb, tt.fields.indent, 1)
			gotN, err := w.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("Write() gotN = %v, want %v", gotN, tt.wantN)
			}
			w.Close()
			gotStr := sb.String()
			if gotStr != tt.wantStr {
				t.Errorf("Write() gotStr = `%v`, want `%v`", gotStr, tt.wantStr)
			}
		})
	}
}

func TestIndenter_IndentBy(t *testing.T) {
	sb := strings.Builder{}
	ind := NewIndenter(&sb, []byte("  "), 0)
	_, _ = ind.WriteString("[\n")
	ind.IndentBy(1)
	_, _ = ind.WriteString("{request:(a b c),device:(d e f)}\n")
	ind.UnindentBy(1)
	_ = ind.WriteByte(']')
	_ = ind.Close()

	{
		const expected = `[
  {request:(a b c),device:(d e f)}
]`
		if actual := sb.String(); actual != expected {
			t.Errorf("UnindentBy() actual = `%s`, want `%s`", actual, expected)
		}
	}
}
