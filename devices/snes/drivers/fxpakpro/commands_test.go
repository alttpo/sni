package fxpakpro

import (
	"context"
	"io"
	"sni/devices"
	"testing"
)

type commandsMock struct {
	t testing.TB
	n int

	next     func(c *commandsMock)
	teardown func(c *commandsMock)

	infoMock func(ctx context.Context) (version, device, rom string, err error)

	resetSystemMock func(ctx context.Context) (err error)
	resetToMenuMock func(ctx context.Context) (err error)
	bootMock        func(ctx context.Context, path string) (err error)

	mkdirMock     func(ctx context.Context, path string) (err error)
	listFilesMock func(ctx context.Context, path string) (files []devices.DirEntry, err error)
	mvMock        func(ctx context.Context, path, newFilename string) (err error)
	rmMock        func(ctx context.Context, path string) (err error)

	getFileMock func(ctx context.Context, path string, w io.Writer, sizeReceived devices.SizeReceivedFunc, progress devices.ProgressReportFunc) (received uint32, err error)
	putFileMock func(ctx context.Context, path string, size uint32, r io.Reader, progress devices.ProgressReportFunc) (n uint32, err error)

	getMock  func(ctx context.Context, space space, address uint32, size uint32) (data []byte, err error)
	vgetMock func(ctx context.Context, space space, chunks ...vgetChunk) (err error)

	putMock  func(ctx context.Context, space space, address uint32, data []byte) (err error)
	vputMock func(ctx context.Context, space space, chunks ...vputChunk) (err error)
}

func (c *commandsMock) advance() {
	if c.next != nil {
		c.next(c)
	}
	c.n++
}

func (c *commandsMock) Close() error {
	return nil
}

func (c *commandsMock) info(ctx context.Context) (version, device, rom string, err error) {
	c.advance()
	if c.infoMock != nil {
		return c.infoMock(ctx)
	}
	return "", "", "", nil
}

func (c *commandsMock) resetSystem(ctx context.Context) (err error) {
	c.advance()
	if c.resetSystemMock != nil {
		return c.resetSystemMock(ctx)
	}
	return nil
}

func (c *commandsMock) resetToMenu(ctx context.Context) (err error) {
	c.advance()
	if c.resetToMenuMock != nil {
		return c.resetToMenuMock(ctx)
	}
	return nil
}

func (c *commandsMock) boot(ctx context.Context, path string) (err error) {
	c.advance()
	if c.bootMock != nil {
		return c.bootMock(ctx, path)
	}
	return nil
}

func (c *commandsMock) mkdir(ctx context.Context, path string) (err error) {
	c.advance()
	if c.mkdirMock != nil {
		return c.mkdirMock(ctx, path)
	}
	return nil
}

func (c *commandsMock) listFiles(ctx context.Context, path string) (files []devices.DirEntry, err error) {
	c.advance()
	if c.listFilesMock != nil {
		return c.listFilesMock(ctx, path)
	}
	return nil, nil
}

func (c *commandsMock) mv(ctx context.Context, path, newFilename string) (err error) {
	c.advance()
	if c.mvMock != nil {
		return c.mvMock(ctx, path, newFilename)
	}
	return nil
}

func (c *commandsMock) rm(ctx context.Context, path string) (err error) {
	c.advance()
	if c.rmMock != nil {
		return c.rmMock(ctx, path)
	}
	return nil
}

func (c *commandsMock) getFile(ctx context.Context, path string, w io.Writer, sizeReceived devices.SizeReceivedFunc, progress devices.ProgressReportFunc) (received uint32, err error) {
	c.advance()
	if c.getFileMock != nil {
		return c.getFileMock(ctx, path, w, sizeReceived, progress)
	}
	return 0, nil
}

func (c *commandsMock) putFile(ctx context.Context, path string, size uint32, r io.Reader, progress devices.ProgressReportFunc) (n uint32, err error) {
	c.advance()
	if c.putFileMock != nil {
		return c.putFileMock(ctx, path, size, r, progress)
	}
	return 0, nil
}

func (c *commandsMock) get(ctx context.Context, space space, address uint32, size uint32) (data []byte, err error) {
	c.advance()
	if c.getMock != nil {
		return c.getMock(ctx, space, address, size)
	}
	return nil, nil
}

func (c *commandsMock) vget(ctx context.Context, space space, chunks ...vgetChunk) (err error) {
	c.advance()
	if c.vgetMock != nil {
		return c.vgetMock(ctx, space, chunks...)
	}
	return nil
}

func (c *commandsMock) put(ctx context.Context, space space, address uint32, data []byte) (err error) {
	c.advance()
	if c.putMock != nil {
		return c.putMock(ctx, space, address, data)
	}
	return nil
}

func (c *commandsMock) vput(ctx context.Context, space space, chunks ...vputChunk) (err error) {
	c.advance()
	if c.vputMock != nil {
		return c.vputMock(ctx, space, chunks...)
	}
	return nil
}
