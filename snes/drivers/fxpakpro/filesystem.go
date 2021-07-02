package fxpakpro

import (
	"context"
	"io"
	"sni/snes"
)

func (d *Device) ReadDirectory(ctx context.Context, path string) ([]snes.DirEntry, error) {
	return d.listFiles(ctx, path)
}

func (d *Device) MakeDirectory(ctx context.Context, path string) error {
	return d.mkdir(ctx, path)
}

func (d *Device) RemoveFile(ctx context.Context, path string) error {
	return d.rm(ctx, path)
}

func (d *Device) RenameFile(ctx context.Context, path, newFilename string) error {
	return d.mv(ctx, path, newFilename)
}

func (d *Device) PutFile(ctx context.Context, path string, r io.Reader, progress snes.ProgressReportFunc) (n uint64, err error) {
	var data []byte
	data, err = io.ReadAll(r)
	if err != nil {
		return
	}

	// TODO: pass `r` into putFile to avoid large allocations
	err = d.putFile(ctx, putFileRequest{
		path:   path,
		rom:    data,
		report: progress,
	})
	if err != nil {
		return
	}

	n = uint64(len(data))
	return
}

func (d *Device) GetFile(ctx context.Context, path string, w io.Writer, progress snes.ProgressReportFunc) (n uint64, err error) {
	n, err = d.getFile(ctx, path, w, progress)
	return
}

func (d *Device) BootFile(ctx context.Context, path string) error {
	return d.boot(ctx, path)
}
