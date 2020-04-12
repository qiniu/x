package bufiox

import (
	"bufio"
	"io"
)

// -----------------------------------------------------------------------------

// Reader class.
type Reader struct {
	bufio.Reader
}

const (
	defaultBufSize = 4096
)

// NewReader returns a new Reader whose buffer has the default size.
func NewReader(rd io.ReadSeeker) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

// NewReaderSize returns a new Reader whose buffer has at least the specified
// size. If the argument io.Reader is already a Reader with large enough size,
// it returns the underlying Reader.
//
func NewReaderSize(rd io.ReadSeeker, size int) *Reader {
	b, ok := rd.(*Reader)
	if ok && b.Size() >= size {
		return b
	}
	r := bufio.NewReaderSize(rd, size)
	return &Reader{Reader: *r}
}

// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence: SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and SeekEnd means
// relative to the end. Seek returns the new offset relative to the start
// of the file and an error, if any.
//
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	return Seek(&r.Reader, offset, whence)
}

// ReadAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// ReadAtLeast returns ErrUnexpectedEOF.
// If min is greater than the length of buf, ReadAtLeast returns ErrShortBuffer.
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.
func (r *Reader) ReadAtLeast(buf []byte, min int) (n int, err error) {
	return ReadAtLeast(&r.Reader, buf, min)
}

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns ErrUnexpectedEOF.
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.
func (r *Reader) ReadFull(buf []byte) (n int, err error) {
	return ReadAtLeast(&r.Reader, buf, len(buf))
}

// -----------------------------------------------------------------------------

// UnderlyingReader returns the underlying reader.
func UnderlyingReader(b interface{}) io.Reader {
	switch v := b.(type) {
	case *Reader:
		return getUnderlyingReader(&v.Reader)
	case *bufio.Reader:
		return getUnderlyingReader(v)
	default:
		panic("can only get the underlying reader of *bufiox.Reader or *bufio.Reader")
	}
}

// -----------------------------------------------------------------------------
