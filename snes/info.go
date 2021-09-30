package snes

import "context"

type Field int

const (
	Field_DeviceName Field = iota
	Field_DeviceVersion
	Field_DeviceStatus
	Field_CoreName
	Field_RomFileName
	Field_RomCRC32
)

type DeviceInfo interface {
	FetchFields(ctx context.Context, fields ...Field) (values []string, err error)
}
