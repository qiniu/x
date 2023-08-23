package fs

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------------------

type unionFS struct {
	fs []http.FileSystem
}

func (p *unionFS) Open(name string) (f http.File, err error) {
	for _, fs := range p.fs {
		f, err = fs.Open(name)
		if !os.IsNotExist(err) {
			return
		}
	}
	return nil, os.ErrNotExist
}

func Union(fs ...http.FileSystem) http.FileSystem {
	return &unionFS{fs}
}

// -----------------------------------------------------------------------------------------

type filesFS struct {
	files []string
}

func (p *filesFS) Open(name string) (f http.File, err error) {
	files := p.files
	name = name[1:]
	for i := 0; i < len(files); i += 2 {
		if files[i] == name {
			f, err = os.Open(files[i+1])
			return
		}
	}
	return nil, os.ErrNotExist
}

func Files(files ...string) http.FileSystem {
	return &filesFS{files}
}

// -----------------------------------------------------------------------------------------

type filesDataFS struct {
	files []string
}

type dataFileInfo struct {
	r    *strings.Reader
	name string
}

func (p *dataFileInfo) Name() string {
	return path.Base(p.name)
}

func (p *dataFileInfo) Size() int64 {
	return p.r.Size()
}

func (p *dataFileInfo) Mode() fs.FileMode {
	return 0
}

func (p *dataFileInfo) ModTime() time.Time {
	return time.Now()
}

func (p *dataFileInfo) IsDir() bool {
	return false
}

func (p *dataFileInfo) Sys() interface{} {
	return nil
}

type dataFile struct {
	*strings.Reader
	name string
}

func (p *dataFile) Close() error {
	return nil
}

func (p *dataFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (p *dataFile) Stat() (fs.FileInfo, error) {
	return &dataFileInfo{p.Reader, p.name}, nil
}

func (p *filesDataFS) Open(name string) (f http.File, err error) {
	files := p.files
	name = name[1:]
	for i := 0; i < len(files); i += 2 {
		if files[i] == name {
			return &dataFile{strings.NewReader(files[i+1]), name}, nil
		}
	}
	return nil, os.ErrNotExist
}

func FilesWithContent(files ...string) http.FileSystem {
	return &filesDataFS{files}
}

// -----------------------------------------------------------------------------------------
