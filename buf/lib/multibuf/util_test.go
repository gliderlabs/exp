package multibuf

import (
	"bytes"
	"testing"
	"time"
)

var (
	lineOfText = []byte("some text\n")
	shortTime  = 100 * time.Millisecond
	longTime   = 500 * time.Millisecond
)

func linesOfText(n int) []byte {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		buf.Write(lineOfText)
	}
	return buf.Bytes()
}

func assertWrite(t *testing.T, N, n int, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if N != n {
		t.Fatalf("unexpected written bytes, expected '%#v' got '%#v' in %s", N, n, t.Name())
	}
}

func waitForWriters(t *testing.T, n int, mwc *MulticastWriteCloser) {
	timeout := time.After(longTime)
	interval := time.After(shortTime)
	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for writers in", t.Name())
		case <-interval:
			if mwc.Count() == n {
				return
			}
		}

	}
}

func waitForDone(t *testing.T, done <-chan bool, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		t.Fatal("timeout waiting for done in", t.Name())
	case <-done:
		return
	}
}

func testWriterCount(t *testing.T, N int, mwc *MulticastWriteCloser) {
	if got := mwc.Count(); got != N {
		t.Fatalf("got writer count '%#v', wanted '%#v' in %s", got, N, t.Name())
	}
}

func testBufferEqual(t *testing.T, buf *bytes.Buffer, b []byte) {
	if got := buf.Bytes(); !bytes.Equal(got, b) {
		t.Fatalf("got buffer bytes '%s', wanted '%s' in %s", got, b, t.Name())
	}
}
