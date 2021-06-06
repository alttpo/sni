package util

import (
	"encoding/hex"
	"encoding/json"
	"strings"
)

type HexBytes []byte

func (b *HexBytes) UnmarshalJSON(j []byte) (err error) {
	var s string
	err = json.Unmarshal(j, &s)
	if err != nil {
		return
	}
	// TODO: trim out comments starting with ';' up to '\n', e.g. "A9 03 ; LDA #$03"
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	*b, err = hex.DecodeString(s)
	return
}
