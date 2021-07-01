package snes

import (
	"context"
	"sni/protos/sni"
)

type DirEntry struct {
	Name string
	Type sni.DirEntryType
	Size uint64
}

type DeviceFilesystem interface {
	ReadDirectory(ctx context.Context, path string) ([]DirEntry, error)
}
