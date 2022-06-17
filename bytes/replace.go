/*
 Copyright 2020 Qiniu Limited (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package bytes

import (
	"bytes"
)

// ---------------------------------------------------

// ReplaceAt does a replace operation at position `off` in place.
func ReplaceAt(b []byte, off, nsrc int, dest []byte) []byte {
	ndelta := len(dest) - nsrc
	if ndelta < 0 {
		left := b[off+nsrc:]
		off += copy(b[off:], dest)
		off += copy(b[off:], left)
		return b[:off]
	}

	if ndelta > 0 {
		b = append(b, dest[:ndelta]...)
		copy(b[off+len(dest):], b[off+nsrc:])
		copy(b[off:], dest)
	} else {
		copy(b[off:], dest)
	}
	return b
}

// ReplaceOne does a replace operation from `from` position in place.
// It returns an offset for next replace operation.
// If the returned offset is -1, it means no replace operation occurred.
func ReplaceOne(b []byte, from int, src, dest []byte) ([]byte, int) {
	pos := bytes.Index(b[from:], src)
	if pos < 0 {
		return b, -1
	}

	from += pos
	return ReplaceAt(b, from, len(src), dest), from + len(dest)
}

// Unlike bytes.Replace, this Replace does replace operations in place.
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
