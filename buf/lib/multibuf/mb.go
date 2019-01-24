package multibuf

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

func NewMulticastBuffer() MulticastBuffer {
	buf := &bytes.Buffer{}
	return &multicastBuffer{
		buffer:   buf,
		reader:   bufio.NewReader(buf),
		peekers:  NewMulticastWriteCloser(),
		flushers: NewMulticastWriteCloser(),
	}
}

type multicastBuffer struct {
	sync.Mutex
	flushers *MulticastWriteCloser
	peekers  *MulticastWriteCloser
	buffer   *bytes.Buffer
	reader   *bufio.Reader
	size     int
	closed   bool
}

func (mb *multicastBuffer) Write(p []byte) (n int, err error) {
	mb.Lock()
	defer mb.Unlock()
	mb.peekers.Write(p)
	if mb.flushers.Count() > 0 {
		return mb.flushers.Write(p)
	}
	n, err = mb.buffer.Write(p)
	mb.size += n
	return
}

func (mb *multicastBuffer) WriteTo(w io.Writer) (n int64, err error) {
	mb.Lock()
	switch w.(type) {
	case Peeker:
		// if data is buffered, Write to peeker.
		// use reader Peek with explicit size to avoid flushing buffer
		if data, err := mb.reader.Peek(mb.size); err == nil {
			nn, err := w.Write(data)
			if err != nil {
				mb.Unlock()
				return int64(nn), err
			}
			n += int64(nn)
		}
		mb.Unlock()
		var nn int64
		nn, err = mb.peekers.WriteTo(w)
		n += nn
		return
	default:
		// if first flusher, WriteTo flusher
		if mb.flushers.Count() == 0 {
			nn, err := mb.reader.WriteTo(w)
			mb.size -= int(nn)
			if err != nil {
				mb.Unlock()
				return nn, err
			}
			n += nn
		}
		mb.Unlock()
		var nn int64
		nn, err = mb.flushers.WriteTo(w)
		n += nn
		return
	}
}

func (mb *multicastBuffer) Close() error {
	mb.Lock()
	defer mb.Unlock()
	mb.flushers.Close()
	mb.peekers.Close()
	mb.closed = true
	return nil
}

func (mb *multicastBuffer) Buffered() int {
	mb.Lock()
	defer mb.Unlock()
	return mb.size
}

func (mb *multicastBuffer) Closed() bool {
	mb.Lock()
	defer mb.Unlock()
	return mb.closed
}
