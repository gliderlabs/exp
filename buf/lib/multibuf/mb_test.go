package multibuf

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestMB_Buffer(t *testing.T) {
	mb := NewMulticastBuffer()

	n, err := mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)
	if got := mb.Buffered(); got != len(lineOfText) {
		t.Fatalf("got buffered len '%#v', wanted '%#v'", got, len(lineOfText))
	}

	buf := &bytes.Buffer{}
	done := make(chan bool)
	go func() {
		mb.WriteTo(buf)
		done <- true
	}()

	mb.Close()
	waitForDone(t, done, shortTime)
	testBufferEqual(t, buf, lineOfText)
}

func TestMB_BufferWithPeeker(t *testing.T) {
	mb := NewMulticastBuffer()

	peeker := &bytes.Buffer{}
	go mb.WriteTo(Peeker{peeker})

	time.Sleep(shortTime)
	n, err := mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)
	if got := mb.Buffered(); got != len(lineOfText) {
		t.Fatalf("got buffered len '%#v', wanted '%#v'", got, len(lineOfText))
	}

	buf := &bytes.Buffer{}
	done := make(chan bool)
	go func() {
		mb.WriteTo(buf)
		done <- true
	}()

	mb.Close()
	waitForDone(t, done, shortTime)
	testBufferEqual(t, buf, lineOfText)
	testBufferEqual(t, peeker, lineOfText)
}

func TestMB_PeekerAndFlusher(t *testing.T) {
	mb := NewMulticastBuffer()

	peeker := &bytes.Buffer{}
	go mb.WriteTo(Peeker{peeker})

	flusher := &bytes.Buffer{}
	flushed := make(chan bool)
	go func() {
		mb.WriteTo(Flusher{flusher})
		flushed <- true
	}()

	time.Sleep(shortTime)
	n, err := mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)
	if got := mb.Buffered(); got != 0 {
		t.Fatalf("got buffered len '%#v', wanted '%#v'", got, 0)
	}

	mb.Close()
	waitForDone(t, flushed, shortTime)
	testBufferEqual(t, flusher, lineOfText)
	testBufferEqual(t, peeker, lineOfText)
}

func TestMB_BufferThenPeeker(t *testing.T) {
	mb := NewMulticastBuffer()
	n, err := mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)

	peeker := &bytes.Buffer{}
	done := make(chan bool)
	go func() {
		mb.WriteTo(Peeker{peeker})
		done <- true
	}()

	time.Sleep(shortTime)
	n, err = mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)

	mb.Close()
	waitForDone(t, done, shortTime)
	testBufferEqual(t, peeker, linesOfText(2))
}

func TestMB_AnyPeekerGetsEntireBuffer(t *testing.T) {
	mb := NewMulticastBuffer()
	var wg sync.WaitGroup

	n, err := mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)

	peeker1 := &bytes.Buffer{}
	wg.Add(1)
	go func() {
		mb.WriteTo(Peeker{peeker1})
		wg.Done()
	}()

	time.Sleep(shortTime)
	n, err = mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)

	peeker2 := &bytes.Buffer{}
	wg.Add(1)
	go func() {
		mb.WriteTo(Peeker{peeker2})
		wg.Done()
	}()

	mb.Close()
	wg.Wait()
	testBufferEqual(t, peeker1, linesOfText(2))
	testBufferEqual(t, peeker2, linesOfText(2))
}

func TestMB_BufferThenFlusher(t *testing.T) {
	mb := NewMulticastBuffer()
	n, err := mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)

	flusher := &bytes.Buffer{}
	done := make(chan bool)
	go func() {
		mb.WriteTo(Flusher{flusher})
		done <- true
	}()

	time.Sleep(shortTime)
	n, err = mb.Write(lineOfText)
	assertWrite(t, len(lineOfText), n, err)

	mb.Close()
	waitForDone(t, done, shortTime)
	testBufferEqual(t, flusher, linesOfText(2))
}
