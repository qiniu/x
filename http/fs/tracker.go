package fs

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"
)

// -----------------------------------------------------------------------------------------

type httpFileInfo struct {
	r    *http.Response
	name string
}

func (p *httpFileInfo) Name() string {
	return path.Base(p.name)
}

func (p *httpFileInfo) Size() int64 {
	return p.r.ContentLength
}

func (p *httpFileInfo) Mode() fs.FileMode {
	return 0
}

func (p *httpFileInfo) ModTime() time.Time {
	if lm := p.r.Header.Get("Last-Modified"); lm != "" {
		if t, err := http.ParseTime(lm); err == nil {
			return t
		}
	}
	return time.Now()
}

func (p *httpFileInfo) IsDir() bool {
	return false
}

func (p *httpFileInfo) Sys() interface{} {
	return nil
}

type httpFile struct {
	file io.ReadCloser
	resp *http.Response
	b    bytes.Buffer
	br   *bytes.Reader
	name string
}

func (p *httpFile) Close() error {
	return p.file.Close()
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

func (p *httpFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (p *httpFile) Stat() (fs.FileInfo, error) {
	return &httpFileInfo{p.resp, p.name}, nil
}

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
	return HttpFile(name, resp), nil
}

func WithTracker(fs http.FileSystem, root string, exts ...string) http.FileSystem {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return &fsWithTracker{fs, m, root}
}

// -----------------------------------------------------------------------------------------
