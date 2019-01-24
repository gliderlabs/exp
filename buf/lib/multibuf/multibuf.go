package multibuf

import (
	"fmt"
	"io"
)

var ErrClosed = fmt.Errorf("writer closed")

type Flusher struct {
	io.Writer
}

type Peeker struct {
	io.Writer
}

type WriterTo struct {
	io.WriterTo
}

func (wt WriterTo) Read(p []byte) (n int, err error) {
	return 0, nil
}

type MulticastBuffer interface {
	io.WriteCloser
	io.WriterTo
	Buffered() int
	Closed() bool
}
