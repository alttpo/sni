package fxpakpro

import (
	"context"
	"sni/snes"
)

func (d *Device) ReadDirectory(ctx context.Context, path string) ([]snes.DirEntry, error) {
	return d.listFiles(ctx, path)
}
