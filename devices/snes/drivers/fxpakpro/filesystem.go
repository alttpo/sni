package fxpakpro

import (
	"context"
	"io"
	"sni/devices"
)

func (d *Device) ReadDirectory(ctx context.Context, path string) ([]devices.DirEntry, error) {
	return d.c.listFiles(ctx, path)
}

func (d *Device) MakeDirectory(ctx context.Context, path string) error {
	return d.c.mkdir(ctx, path)
}

func (d *Device) RemoveFile(ctx context.Context, path string) error {
	return d.c.rm(ctx, path)
}

func (d *Device) RenameFile(ctx context.Context, path, newFilename string) error {
	return d.c.mv(ctx, path, newFilename)
}

func (d *Device) PutFile(ctx context.Context, path string, size uint32, r io.Reader, progress devices.ProgressReportFunc) (n uint32, err error) {
	n, err = d.c.putFile(ctx, path, size, r, progress)
	return
}

func (d *Device) GetFile(ctx context.Context, path string, w io.Writer, sizeReceived devices.SizeReceivedFunc, progress devices.ProgressReportFunc) (size uint32, err error) {
	size, err = d.c.getFile(ctx, path, w, sizeReceived, progress)
	return
}

func (d *Device) BootFile(ctx context.Context, path string) error {
	return d.c.boot(ctx, path)
}
