package hex

import (
	"strings"
	"testing"
)

func TestDumper(t *testing.T) {
	sb := &strings.Builder{}
	hd := Dumper(sb, uint(0xf50010))
	_, _ = hd.Write([]byte{0x07, 0x09, 0x0a, 0x55, 0xaa})
	_ = hd.Close()

	{
		const expected = `00f50010  07 09 0a 55 aa                                    |...U.|
`
		if actual := sb.String(); actual != expected {
			t.Errorf("Dumper() actual = `%v`, want `%v`", actual, expected)
		}
	}
}
