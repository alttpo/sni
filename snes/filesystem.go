package snes

import (
	"context"
	"io"
	"sni/protos/sni"
)

type DirEntry struct {
	Name string
	Type sni.DirEntryType
}

type ProgressReportFunc func(current uint64, total uint64)

type DeviceFilesystem interface {
	ReadDirectory(ctx context.Context, path string) ([]DirEntry, error)
}
