package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"path"

	"github.com/qiniu/x/http/fs"
)

type fsGzip struct {
	fs   http.FileSystem
	exts map[string]struct{}
}

func (p *fsGzip) Open(name string) (file http.File, err error) {
	file, err = p.fs.Open(name)
	if err != nil {
		return
	}
	ext := path.Ext(name)
	if _, ok := p.exts[ext]; !ok {
		return
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	defer gr.Close()
	b, err := io.ReadAll(gr)
	if err != nil {
		return
	}
	return fs.File(name, bytes.NewReader(b)), nil
}

func New(fs http.FileSystem, exts ...string) http.FileSystem {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return &fsGzip{fs, m}
}
