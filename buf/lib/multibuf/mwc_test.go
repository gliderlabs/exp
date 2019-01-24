package multibuf

import (
	"bytes"
	"testing"
	"time"
)

func TestMWC_Write(t *testing.T) {
	mwc := NewMulticastWriteCloser()
	defer mwc.Close()

	out := &bytes.Buffer{}
	in := linesOfText(3)

	go mwc.WriteTo(out)
	waitForWriters(t, 1, mwc)

	n, err := mwc.Write(in)
	assertWrite(t, len(in), n, err)

	testBufferEqual(t, out, in)
}

func TestMWC_NoWriteAfterClose(t *testing.T) {
	mwc := NewMulticastWriteCloser()
	mwc.Close()
	_, err := mwc.Write(lineOfText)
	if err == nil {
		t.Fatal("write allowed after close")
	}
}

func TestMWC_MulticastWrite(t *testing.T) {
	mwc := NewMulticastWriteCloser()
	defer mwc.Close()

	out1 := &bytes.Buffer{}
	out2 := &bytes.Buffer{}
	out3 := &bytes.Buffer{}
	in := linesOfText(3)

	go mwc.WriteTo(out1)
	go mwc.WriteTo(out2)
	go mwc.WriteTo(out3)
	waitForWriters(t, 3, mwc)

	n, err := mwc.Write(in)
	assertWrite(t, len(in), n, err)

	testBufferEqual(t, out1, in)
	testBufferEqual(t, out2, in)
	testBufferEqual(t, out3, in)
}

func TestMWC_WriteToBlocksUntilClose(t *testing.T) {
	mwc := NewMulticastWriteCloser()
	out := &bytes.Buffer{}
	in := linesOfText(3)

	done := make(chan bool)
	go func() {
		mwc.WriteTo(out)
		done <- true
	}()
	waitForWriters(t, 1, mwc)

	n, err := mwc.Write(in)
	assertWrite(t, len(in), n, err)
	testWriterCount(t, 1, mwc)

	mwc.Close()
	select {
	case <-time.After(shortTime):
		t.Fatal("timeout waiting for done in", t.Name())
	case <-done:
		testWriterCount(t, 0, mwc)
		return
	}
}
