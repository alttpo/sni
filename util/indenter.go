package util

import (
	"bytes"
	"io"
)

// Indenter to indent line by line output, buffering the current line until the next newline is seen
// or Close() is called.
type Indenter struct {
	o     io.Writer
	tab   []byte
	depth int

	line []byte
}

// NewIndenter creates an Indenter that writes to the given writer and initializes the indentation
// character sequence as well as the indentation depth.
func NewIndenter(writer io.Writer, tab []byte, depth int) *Indenter {
	return &Indenter{
		o:     writer,
		tab:   tab,
		depth: depth,
	}
}

// IndentBy affects the indentation of the current buffered line and subsequent lines
func (w *Indenter) IndentBy(n int) {
	w.depth += n
}

// UnindentBy affects the indentation of the current buffered line and subsequent lines
func (w *Indenter) UnindentBy(n int) {
	w.depth -= n
}

func (w *Indenter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

func (w *Indenter) WriteByte(c byte) (err error) {
	_, err = w.Write([]byte{c})
	return
}

func (w *Indenter) Write(p []byte) (n int, err error) {
	n = 0
	err = nil

	d := p
	for {
		i := bytes.IndexByte(d, '\n')
		if i < 0 {
			// no newline found; buffer the running line contents:
			w.line = append(w.line, d...)
			n, err = len(p), nil
			return
		}
		i++

		// append up to and including the newline:
		w.line = append(w.line, d[:i]...)

		err = w.writeLine()
		if err != nil {
			return
		}

		// clear running line contents:
		w.line = w.line[:0]

		d = d[i:]
	}
}

func (w *Indenter) writeLine() (err error) {
	// prepend line with indentation tabs:
	buf := make([]byte, 0, len(w.line)+len(w.tab)*w.depth)
	for c := 0; c < w.depth; c++ {
		buf = append(buf, w.tab...)
	}
	buf = append(buf, w.line...)

	// emit the line:
	_, err = w.o.Write(buf)
	return
}

func (w *Indenter) Close() (err error) {
	if len(w.line) == 0 {
		return
	}

	return w.writeLine()
}
