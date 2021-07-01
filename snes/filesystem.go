package snes

import (
	"context"
	"sni/protos/sni"
)

type DirEntry struct {
	Name string
	Type sni.DirEntryType
}

type DeviceFilesystem interface {
	ReadDirectory(ctx context.Context, path string) ([]DirEntry, error)
}
