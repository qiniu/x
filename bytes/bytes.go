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
	"io"
	"syscall"
)

// ---------------------------------------------------

// Unlike the standard library's bytes.Reader, this Reader supports Seek.
type Reader struct {
	b   []byte
	off int
}

// NewReader create a readonly stream for byte slice:
//
//	 var slice []byte
//	 ...
//	 r := bytes.NewReader(slice)
//	 ...
//	 r.Seek(0, 0) // r.SeekToBegin()
//	 ...
//
// Unlike the standard library's bytes.Reader, the returned Reader supports Seek.
func NewReader(val []byte) *Reader {
	return &Reader{val, 0}
}

func (r *Reader) Len() int {
	if r.off >= len(r.b) {
		return 0
	}
	return len(r.b) - r.off
}

func (r *Reader) Bytes() []byte {
	return r.b[r.off:]
}

func (r *Reader) SeekToBegin() (err error) {
	r.off = 0
	return
}

func (r *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
	case 1:
		offset += int64(r.off)
	case 2:
		offset += int64(len(r.b))
	default:
		err = syscall.EINVAL
		return
	}
	if offset < 0 {
		err = syscall.EINVAL
		return
	}
	if offset >= int64(len(r.b)) {
		r.off = len(r.b)
	} else {
		r.off = int(offset)
	}
	ret = int64(r.off)
	return
}

func (r *Reader) Read(val []byte) (n int, err error) {
	n = copy(val, r.b[r.off:])
	if n == 0 && len(val) != 0 {
		err = io.EOF
		return
	}
	r.off += n
	return
}

func (r *Reader) Close() (err error) {
	return
}

// ---------------------------------------------------

// Writer implements a write stream with a limited capacity.
type Writer struct {
	b []byte
	n int
}

// NewWriter NewWriter creates a write stream with a limited capacity:
//
//	 slice := make([]byte, 1024)
//	 w := bytes.NewWriter(slice)
//	 ...
//	 writtenData := w.Bytes()
//
// If we write more than 1024 bytes of data to w, the excess data will be discarded.
func NewWriter(buff []byte) *Writer {
	return &Writer{buff, 0}
}

func (p *Writer) Write(val []byte) (n int, err error) {
	n = copy(p.b[p.n:], val)
	if n == 0 && len(val) > 0 {
		err = io.EOF
		return
	}
	p.n += n
	return
}

func (p *Writer) Len() int {
	return p.n
}

func (p *Writer) Bytes() []byte {
	return p.b[:p.n]
}

// Reset deletes all written data.
func (p *Writer) Reset() {
	p.n = 0
}

// ---------------------------------------------------

type Buffer struct {
	b []byte
}

// NewBuffer creates a random read/write memory file that supports ReadAt/WriteAt
// methods instead of Read/Write:
//
//	 b := bytes.NewBuffer()
//	 b.Truncate(100)
//	 b.WriteAt([]byte("hello"), 100)
//	 slice := make([]byte, 105)
//	 n, err := b.ReadAt(slice, 0)
//	 ...
//
func NewBuffer() *Buffer {
	return new(Buffer)
}

func (p *Buffer) ReadAt(buf []byte, off int64) (n int, err error) {
	ioff := int(off)
	if len(p.b) <= ioff {
		return 0, io.EOF
	}
	n = copy(buf, p.b[ioff:])
	if n != len(buf) {
		err = io.EOF
	}
	return
}

func (p *Buffer) WriteAt(buf []byte, off int64) (n int, err error) {
	ioff := int(off)
	iend := ioff + len(buf)
	if len(p.b) < iend {
		if len(p.b) == ioff {
			p.b = append(p.b, buf...)
			return len(buf), nil
		}
		zero := make([]byte, iend-len(p.b))
		p.b = append(p.b, zero...)
	}
	copy(p.b[ioff:], buf)
	return len(buf), nil
}

func (p *Buffer) WriteStringAt(buf string, off int64) (n int, err error) {
	ioff := int(off)
	iend := ioff + len(buf)
	if len(p.b) < iend {
		if len(p.b) == ioff {
			p.b = append(p.b, buf...)
			return len(buf), nil
		}
		zero := make([]byte, iend-len(p.b))
		p.b = append(p.b, zero...)
	}
	copy(p.b[ioff:], buf)
	return len(buf), nil
}

func (p *Buffer) Truncate(fsize int64) (err error) {
	size := int(fsize)
	if len(p.b) < size {
		zero := make([]byte, size-len(p.b))
		p.b = append(p.b, zero...)
	} else {
		p.b = p.b[:size]
	}
	return nil
}

func (p *Buffer) Buffer() []byte {
	return p.b
}

func (p *Buffer) Len() int {
	return len(p.b)
}

// ---------------------------------------------------
