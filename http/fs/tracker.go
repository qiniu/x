package fs

import (
	"bytes"
	"io"
	"net/http"
	"path"
	"time"
)

// -----------------------------------------------------------------------------------------

type httpContent struct {
	file io.ReadCloser
	resp *http.Response
	b    bytes.Buffer
	br   *bytes.Reader
}

func (p *httpContent) Read(b []byte) (n int, err error) {
	if p.br == nil {
		n, err = p.file.Read(b)
		p.b.Write(b[:n])
	} else {
		n, err = p.br.Read(b)
	}
	return
}

func (p *httpContent) Seek(offset int64, whence int) (int64, error) {
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

func (p *httpContent) Size() int64 {
	return p.resp.ContentLength
}

func (p *httpContent) Close() error {
	return p.file.Close()
}

func (p *httpContent) ModTime() time.Time {
	if lm := p.resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := http.ParseTime(lm); err == nil {
			return t
		}
	}
	return time.Now()
}

// HttpFile implements a http.File by a http.Response object.
func HttpFile(name string, resp *http.Response) http.File {
	return File(name, &httpContent{file: resp.Body, resp: resp})
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

// WithTracker implements a http.FileSystem by pactching large file access like git lfs.
func WithTracker(fs http.FileSystem, root string, exts ...string) http.FileSystem {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return &fsWithTracker{fs, m, root}
}

// -----------------------------------------------------------------------------------------
