package bufiox

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unsafe"
)

// -----------------------------------------------------------------------------

type nilReaderImpl int

func (p nilReaderImpl) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

var nilReader io.Reader = nilReaderImpl(0)

// -----------------------------------------------------------------------------

type reader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int
	lastRuneSize int
}

// NewReaderBuffer returns a new Reader who uses a extern buffer.
func NewReaderBuffer(buf []byte) *bufio.Reader {
	r := &reader{
		buf:          buf,
		rd:           nilReader,
		w:            len(buf),
		lastByte:     -1,
		lastRuneSize: -1,
	}
	b := new(bufio.Reader)
	*b = *(*bufio.Reader)(unsafe.Pointer(r))
	return b
}

// Buffer is reserved for internal use.
func Buffer(b *bufio.Reader) []byte {
	r := (*reader)(unsafe.Pointer(b))
	return r.buf
}

// IsReaderBuffer returns if this Reader instance is returned by NewReaderBuffer
func IsReaderBuffer(b *bufio.Reader) bool {
	r := (*reader)(unsafe.Pointer(b))
	return r.rd == nilReader
}

// ReadAll reads all data
func ReadAll(b *bufio.Reader) (ret []byte, err error) {
	r := (*reader)(unsafe.Pointer(b))
	if r.rd == nilReader {
		ret, r.buf = r.buf, nil
		return
	}
	var w bytes.Buffer
	_, err1 := b.WriteTo(&w)
	return w.Bytes(), err1
}

// -----------------------------------------------------------------------------

func getUnderlyingReader(b *bufio.Reader) io.Reader {
	r := (*reader)(unsafe.Pointer(b))
	return r.rd
}

// ErrSeekUnsupported error.
var ErrSeekUnsupported = errors.New("bufio: the underlying reader doesn't support seek")

// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence: SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and SeekEnd means
// relative to the end. Seek returns the new offset relative to the start
// of the file and an error, if any.
//
func Seek(b *bufio.Reader, offset int64, whence int) (int64, error) {
	r := getUnderlyingReader(b)
	if seeker, ok := r.(io.Seeker); ok {
		newoff, err := seeker.Seek(offset, whence)
		if err == nil {
			b.Reset(r)
		}
		return newoff, err
	}
	return 0, ErrSeekUnsupported
}

// ReadAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// ReadAtLeast returns ErrUnexpectedEOF.
// If min is greater than the length of buf, ReadAtLeast returns ErrShortBuffer.
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.
func ReadAtLeast(r *bufio.Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, io.ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns ErrUnexpectedEOF.
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.
func ReadFull(r *bufio.Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}

// -----------------------------------------------------------------------------
