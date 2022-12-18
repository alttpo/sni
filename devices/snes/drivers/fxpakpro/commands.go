package fxpakpro

import (
	"context"
	"io"
	"sni/devices"
)

type commands interface {
	io.Closer
	info(ctx context.Context) (version, device, rom string, err error)
	resetSystem(ctx context.Context) (err error)
	resetToMenu(ctx context.Context) (err error)
	mkdir(ctx context.Context, path string) (err error)
	listFiles(ctx context.Context, path string) (files []devices.DirEntry, err error)
	mv(ctx context.Context, path, newFilename string) (err error)
	rm(ctx context.Context, path string) (err error)
	boot(ctx context.Context, path string) (err error)
	get(ctx context.Context, space space, address uint32, size uint32) (data []byte, err error)
	getFile(ctx context.Context, path string, w io.Writer, sizeReceived devices.SizeReceivedFunc, progress devices.ProgressReportFunc) (received uint32, err error)
	vget(ctx context.Context, space space, chunks ...vgetChunk) (err error)
	put(ctx context.Context, space space, address uint32, data []byte) (err error)
	vput(ctx context.Context, space space, chunks ...vputChunk) (err error)
	putFile(ctx context.Context, path string, size uint32, r io.Reader, progress devices.ProgressReportFunc) (n uint32, err error)
}
