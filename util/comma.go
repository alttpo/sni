package util

import "strings"

func Delimited(items []string) string {
	l := len(items)
	sb := strings.Builder{}
	for i, s := range items {
		sb.WriteString(s)
		if i < l-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}

func DelimitedGen(items []interface{}, mapper func(interface{}) string) string {
	l := len(items)
	sb := strings.Builder{}
	for i, v := range items {
		s := mapper(v)
		sb.WriteString(s)
		if i < l-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}
