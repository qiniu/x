package fs

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------------------

type ContentReader interface {
	io.Reader
	io.Seeker
	Size() int64
}

type dataFileInfo struct {
	r    ContentReader
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
	ContentReader
	name string
}

func (p *dataFile) Close() error {
	return nil
}

func (p *dataFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (p *dataFile) Stat() (fs.FileInfo, error) {
	return &dataFileInfo{p.ContentReader, p.name}, nil
}

func File(name string, r ContentReader) http.File {
	return &dataFile{r, name}
}

// -----------------------------------------------------------------------------------------

type filesDataFS struct {
	files []string
}

func (p *filesDataFS) Open(name string) (f http.File, err error) {
	files := p.files
	name = name[1:]
	for i := 0; i < len(files); i += 2 {
		if files[i] == name {
			return File(name, strings.NewReader(files[i+1])), nil
		}
	}
	return nil, os.ErrNotExist
}

func FilesWithContent(files ...string) http.FileSystem {
	return &filesDataFS{files}
}

// -----------------------------------------------------------------------------------------
