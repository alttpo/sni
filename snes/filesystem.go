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

type ProgressReportFunc func(current uint32, total uint32)

type DeviceFilesystem interface {
	ReadDirectory(ctx context.Context, path string) ([]DirEntry, error)
	MakeDirectory(ctx context.Context, path string) error
	RemoveFile(ctx context.Context, path string) error
	RenameFile(ctx context.Context, path, newFilename string) error
	PutFile(ctx context.Context, path string, size uint32, r io.Reader, progress ProgressReportFunc) (n uint32, err error)
	GetFile(ctx context.Context, path string, w io.Writer, progress ProgressReportFunc) (size uint32, err error)
	BootFile(ctx context.Context, path string) error
}
