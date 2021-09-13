package snes

import "context"

type Field int

const (
	Field_DeviceName Field = iota
	Field_DeviceVersion
	Field_RomFileName
)

type DeviceInfo interface {
	FetchFields(ctx context.Context, fields ...Field) (values []string, err error)
}
