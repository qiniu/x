package fs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"time"
)

// -----------------------------------------------------------------------------------------

type stream struct {
	file io.ReadCloser
	b    bytes.Buffer
	br   *bytes.Reader
	name string
}

func (p *stream) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, os.ErrInvalid
}

func (p *stream) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (p *stream) Stat() (fs.FileInfo, error) {
	return &dataFileInfo{p, p.name}, nil
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

func (p *stream) Seek(offset int64, whence int) (int64, error) {
	if p.br == nil {
		off := p.b.Len()
		_, err := io.Copy(&p.b, p.file)
		if err != nil {
			return 0, err
		}
		p.br = bytes.NewReader(p.b.Bytes())
		p.br.Seek(int64(off), io.SeekStart)
	}
	return p.br.Seek(offset, whence)
}

func (p *stream) Size() int64 {
	if r, ok := p.file.(interface{ Size() int64 }); ok {
		return r.Size()
	}
	return -1
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

// SequenceFile implements a http.File by a io.ReadCloser object.
func SequenceFile(name string, body io.ReadCloser) http.File {
	return &stream{file: body}
}

// -----------------------------------------------------------------------------------------

type httpFile struct {
	file io.ReadCloser
	resp *http.Response
	b    bytes.Buffer
	br   *bytes.Reader
	name string
}

// TryReader is provided for fast copy. See github.com/x/fs/cached/lfs.download
func (p *httpFile) TryReader() *bytes.Reader {
	return p.br
}

func (p *httpFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, os.ErrInvalid
}

func (p *httpFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (p *httpFile) Stat() (fs.FileInfo, error) {
	return &dataFileInfo{p, p.name}, nil
}

func (p *httpFile) Read(b []byte) (n int, err error) {
	if p.br == nil {
		n, err = p.file.Read(b)
		p.b.Write(b[:n])
	} else {
		n, err = p.br.Read(b)
	}
	return
}

func (p *httpFile) Seek(offset int64, whence int) (int64, error) {
	if p.br == nil {
		off := p.b.Len()
		_, err := io.Copy(&p.b, p.file)
		if err != nil {
			return 0, err
		}
		p.br = bytes.NewReader(p.b.Bytes())
		p.br.Seek(int64(off), io.SeekStart)
	}
	return p.br.Seek(offset, whence)
}

func (p *httpFile) Size() int64 {
	return p.resp.ContentLength
}

func (p *httpFile) Close() error {
	return p.file.Close()
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
	return &httpFile{file: resp.Body, resp: resp, name: name}
}

// -----------------------------------------------------------------------------------------

type fsWithTracker struct {
	fs   http.FileSystem
	exts map[string]struct{}
	root string
}

func (p *fsWithTracker) Open(name string) (file http.File, err error) {
	ext := path.Ext(name)
	if _, ok := p.exts[ext]; !ok {
		return p.fs.Open(name)
	}
	resp, err := http.Get(p.root + name)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		url := "url"
		if req := resp.Request; req != nil {
			url = req.URL.String()
		}
		return nil, fmt.Errorf("http.Get %s error: status %d (%s)", url, resp.StatusCode, resp.Status)
	}
	return HttpFile(name, resp), nil
}

// WithTracker implements a http.FileSystem by pactching large file access like git lfs.
func WithTracker(fs http.FileSystem, root string, exts ...string) http.FileSystem {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return &fsWithTracker{fs, m, root}
}

// -----------------------------------------------------------------------------------------