/*
 Copyright 2023 Qiniu Limited (qiniu.com)

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

package fs

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"time"
)

// Download downloads from a http.File.
func Download(destFile string, src http.File) (err error) {
	f, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer f.Close()

	return CopyFile(f, src)
}

// CopyFile copies from a http.File to a io.Writer.
func CopyFile(f io.Writer, src http.File) (err error) {
	if tr, ok := src.(interface{ TryReader() *bytes.Reader }); ok {
		if r := tr.TryReader(); r != nil {
			_, err = r.WriteTo(f)
			return
		}
	}
	// Unseekable(src): reduce unnecessary memory usage
	_, err = io.Copy(f, Unseekable(src))
	return
}

// -----------------------------------------------------------------------------------------

// Unseekable convert a http.File into a io.ReadCloser object without io.Seeker.
// Note you should stop using the origin http.File object to read data.
// This method is used to reduce unnecessary memory usage.
func Unseekable(file http.File) io.ReadCloser {
	switch f := file.(type) {
	case *stream:
		if f.br == nil {
			return f.file
		}
	case *httpFile:
		if f.br == nil {
			return f.file
		}
	}
	return file
}

// -----------------------------------------------------------------------------------------

type stream struct {
	file io.ReadCloser
	b    bytes.Buffer
	br   *bytes.Reader
	name string
}

// TryReader is provided for fast copy. See function `Download`.
func (p *stream) TryReader() *bytes.Reader {
	return p.br
}

func (p *stream) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, fs.ErrInvalid
}

func (p *stream) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (p *stream) IsDir() bool {
	return false
}

func (p *stream) Mode() fs.FileMode {
	return fs.ModeIrregular
}

func (p *stream) Name() string {
	return path.Base(p.name)
}

func (p *stream) FullName() string {
	return p.name
}

func (p *stream) Stat() (fs.FileInfo, error) {
	return p, nil
}

func (p *stream) Read(b []byte) (n int, err error) {
	if p.br == nil {
		n, err = p.file.Read(b)
		p.b.Write(b[:n])
	} else {
		n, err = p.br.Read(b)
	}
	return
}

func (p *stream) Seek(offset int64, whence int) (_ int64, err error) {
	if p.br == nil {
		if err = p.initBr(); err != nil {
			return
		}
	}
	return p.br.Seek(offset, whence)
}

func (p *stream) initBr() (err error) {
	off := p.b.Len()
	_, err = io.Copy(&p.b, p.file)
	if err == nil {
		p.br = bytes.NewReader(p.b.Bytes())
		p.br.Seek(int64(off), io.SeekStart)
	}
	return
}

func (p *stream) getSize() int64 {
	if p.br == nil {
		if err := p.initBr(); err != nil {
			panic(err)
		}
	}
	return p.br.Size()
}

func (p *stream) Size() int64 {
	if r, ok := p.file.(interface{ Size() int64 }); ok {
		return r.Size()
	}
	return p.getSize()
}

func (p *stream) Close() error {
	return p.file.Close()
}

func (p *stream) ModTime() time.Time {
	if r, ok := p.file.(interface{ ModTime() time.Time }); ok {
		return r.ModTime()
	}
	return time.Now()
}

func (p *stream) Sys() any {
	return nil
}

// SequenceFile implements a http.File by a io.ReadCloser object.
func SequenceFile(name string, body io.ReadCloser) http.File {
	return &stream{name: name, file: body}
}

// -----------------------------------------------------------------------------------------

type httpFile struct {
	stream
	resp *http.Response
}

func (p *httpFile) Size() int64 {
	if n := p.resp.ContentLength; n >= 0 {
		return n
	}
	return p.getSize()
}

func (p *httpFile) ModTime() time.Time {
	if lm := p.resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := http.ParseTime(lm); err == nil {
			return t
		}
	}
	return time.Now()
}

// HttpFile implements a http.File by a http.Response object.
func HttpFile(name string, resp *http.Response) http.File {
	return &httpFile{stream{file: resp.Body, name: name}, resp}
}

// -----------------------------------------------------------------------------------------
