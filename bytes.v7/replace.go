package bytes

import (
	"bytes"
)

// ---------------------------------------------------

func ReplaceAt(b []byte, off int, src, dest []byte) []byte {

	nsrc, ndest := len(src), len(dest)
	if nsrc >= ndest {
		left := b[off+nsrc:]
		off += copy(b[off:], dest)
		off += copy(b[off:], left)
		return b[:off]
	}
	tailoff := len(b) - (ndest - nsrc)
	b = append(b, b[tailoff:]...)
	copy(b[off+ndest:], b[off+nsrc:len(b)])
	copy(b[off:], dest)
	return b
}

func ReplaceOne(b []byte, from int, src, dest []byte) ([]byte, int) {

	pos := bytes.Index(b[from:], src)
	if pos < 0 {
		return b, -1
	}

	from += pos
	return ReplaceAt(b, from, src, dest), from + len(dest)
}

func Replace(b []byte, src, dest []byte, n int) []byte {

	from := 0
	for n != 0 {
		b, from = ReplaceOne(b, from, src, dest)
		if from < 0 {
			break
		}
		n--
	}
	return b
}

// ---------------------------------------------------

