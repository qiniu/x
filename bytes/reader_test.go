package bytes_test

import (
	"io"
	"sync"
	"syscall"
	"testing"

	. "github.com/qiniu/x/bytes"
)

func TestReader(t *testing.T) {
	r := NewReader([]byte("0123456789"))
	tests := []struct {
		off     int64
		seek    int
		n       int
		want    string
		wantpos int64
		readerr error
		seekerr string
	}{
		{seek: io.SeekStart, off: 0, n: 20, want: "0123456789"},
		{seek: io.SeekStart, off: 1, n: 1, want: "1"},
		{seek: io.SeekCurrent, off: 1, wantpos: 3, n: 2, want: "34"},
		{seek: io.SeekStart, off: -1, seekerr: "invalid argument"},
		{seek: io.SeekStart, off: 1 << 33, wantpos: 10, readerr: nil},
		{seek: io.SeekCurrent, off: 1, wantpos: 10, readerr: nil},
		{seek: io.SeekStart, n: 5, want: "01234"},
		{seek: io.SeekCurrent, n: 5, want: "56789"},
		{seek: io.SeekEnd, off: -1, n: 1, wantpos: 9, want: "9"},
	}

	for i, tt := range tests {
		pos, err := r.Seek(tt.off, tt.seek)
		if err == nil && tt.seekerr != "" {
			t.Errorf("%d. want seek error %q", i, tt.seekerr)
			continue
		}
		if err != nil && err.Error() != tt.seekerr {
			t.Errorf("%d. seek error = %q; want %q", i, err.Error(), tt.seekerr)
			continue
		}
		if tt.wantpos != 0 && tt.wantpos != pos {
			t.Errorf("%d. pos = %d, want %d", i, pos, tt.wantpos)
		}
		buf := make([]byte, tt.n)
		n, err := r.Read(buf)
		if err != tt.readerr {
			t.Errorf("%d. read = %v; want %v", i, err, tt.readerr)
			continue
		}
		got := string(buf[:n])
		if got != tt.want {
			t.Errorf("%d. got %q; want %q", i, got, tt.want)
		}
	}
}

func TestReadAfterBigSeek(t *testing.T) {
	r := NewReader([]byte("0123456789"))
	if _, err := r.Seek(1<<31+5, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if n, err := r.Read(make([]byte, 10)); n != 0 || err != io.EOF {
		t.Errorf("Read = %d, %v; want 0, EOF", n, err)
	}
}

func TestEmptyReaderConcurrent(t *testing.T) {
	// Test for the race detector, to verify a Read that doesn't yield any bytes
	// is okay to use from multiple goroutines. This was our historic behavior.
	// See golang.org/issue/7856
	r := NewReader([]byte{})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			var buf [1]byte
			r.Read(buf[:])
		}()
		go func() {
			defer wg.Done()
			r.Read(nil)
		}()
	}
	wg.Wait()
}

func TestReaderLen(t *testing.T) {
	const data = "hello world"
	r := NewReader([]byte(data))
	if got, want := r.Len(), 11; got != want {
		t.Errorf("r.Len(): got %d, want %d", got, want)
	}
	if n, err := r.Read(make([]byte, 10)); err != nil || n != 10 {
		t.Errorf("Read failed: read %d %v", n, err)
	}
	if got, want := r.Len(), 1; got != want {
		t.Errorf("r.Len(): got %d, want %d", got, want)
	}
	if n, err := r.Read(make([]byte, 1)); err != nil || n != 1 {
		t.Errorf("Read failed: read %d %v; want 1, nil", n, err)
	}
	if got, want := r.Len(), 0; got != want {
		t.Errorf("r.Len(): got %d, want %d", got, want)
	}
}

// verify that copying from an empty reader always has the same results,
// regardless of the presence of a WriteTo method.
func TestReaderCopyNothing(t *testing.T) {
	type nErr struct {
		n   int64
		err error
	}
	type justReader struct {
		io.Reader
	}
	type justWriter struct {
		io.Writer
	}
	discard := justWriter{io.Discard} // hide ReadFrom

	var with, withOut nErr
	with.n, with.err = io.Copy(discard, NewReader(nil))
	withOut.n, withOut.err = io.Copy(discard, justReader{NewReader(nil)})
	if with != withOut {
		t.Errorf("behavior differs: with = %#v; without: %#v", with, withOut)
	}
}

// tests that Len is affected by reads, but Size is not.
func TestReaderLenSize(t *testing.T) {
	r := NewReader([]byte("abc"))
	io.CopyN(io.Discard, r, 1)
	if r.Len() != 2 {
		t.Errorf("Len = %d; want 2", r.Len())
	}
	if r.Size() != 3 {
		t.Errorf("Size = %d; want 3", r.Size())
	}
}

func TestReaderZero(t *testing.T) {
	if l := (&Reader{}).Len(); l != 0 {
		t.Errorf("Len: got %d, want 0", l)
	}

	if e := (&Reader{}).SeekToBegin(); e != nil {
		t.Errorf("SeekToBegin: got %v, want nil", e)
	}

	if _, e := (&Reader{}).Seek(0, 4); e != syscall.EINVAL {
		t.Errorf("SeekToBegin: got %v, want EINVAL", e)
	}

	if e := (&Reader{}).Close(); e != nil {
		t.Errorf("Close: got %v, want nil", e)
	}

	if v := (&Reader{}).Bytes(); v != nil {
		t.Errorf("Bytes: got %v, want nil", nil)
	}

	if n, err := (&Reader{}).Read(nil); n != 0 || err != nil {
		t.Errorf("Read: got %d, %v; want 0, io.EOF", n, err)
	}

	if offset, err := (&Reader{}).Seek(11, io.SeekStart); offset != 0 || err != nil {
		t.Errorf("Seek: got %d, %v; want 11, nil", offset, err)
	}

	if s := (&Reader{}).Size(); s != 0 {
		t.Errorf("Size: got %d, want 0", s)
	}
}
