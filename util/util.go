package util

import (
	"strconv"
	"strings"
)

func IsTruthy(v string) bool {
	var err error
	b, err := strconv.ParseBool(v)
	if err == nil {
		return b
	}
	d, err := strconv.ParseInt(v, 10, 32)
	if err == nil {
		return d != 0
	}
	v = strings.ToLower(v)
	if v == "on" {
		return true
	}
	if v == "yes" {
		return true
	}
	if v == "enabled" {
		return true
	}
	return false
}
