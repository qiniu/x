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

// NewReader returns a new Reader whose buffer has the default size.
func NewReader(rd io.ReadSeeker) *Reader {
	r := bufio.NewReader(rd)
	return &Reader{Reader: *r}
}

// NewReaderSize returns a new Reader whose buffer has at least the specified
// size. If the argument io.Reader is already a Reader with large enough size,
// it returns the underlying Reader.
//
func NewReaderSize(rd io.ReadSeeker, size int) *Reader {
	r := bufio.NewReaderSize(rd, size)
	return &Reader{Reader: *r}
}

// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence: SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and SeekEnd means
// relative to the end. Seek returns the new offset relative to the start
// of the file and an error, if any.
//
func (p *Reader) Seek(offset int64, whence int) (int64, error) {
	return Seek(&p.Reader, offset, whence)
}

// -----------------------------------------------------------------------------
