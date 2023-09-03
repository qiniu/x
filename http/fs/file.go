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

type dataFile struct {
	ContentReader
	name string
}

func (p *dataFile) Close() error {
	if r, ok := p.ContentReader.(io.Closer); ok {
		return r.Close()
	}
	return nil
}

func (p *dataFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, os.ErrInvalid
}

func (p *dataFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (p *dataFile) IsDir() bool {
	return false
}

func (p *dataFile) Mode() fs.FileMode {
	return 0
}

func (p *dataFile) Stat() (fs.FileInfo, error) {
	return p, nil
}

func (p *dataFile) Name() string {
	return path.Base(p.name)
}

func (p *dataFile) Size() int64 {
	return p.ContentReader.Size()
}

func (p *dataFile) ModTime() time.Time {
	if r, ok := p.ContentReader.(interface{ ModTime() time.Time }); ok {
		return r.ModTime()
	}
	return time.Now()
}

func (p *dataFile) Sys() interface{} {
	return nil
}

// File implements a http.File by a ContentReader which may implement
// optional interface{ ModTime() time.Time } and io.Closer.
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

// FilesWithContent implements a http.FileSystem by a list of file name and content.
func FilesWithContent(files ...string) http.FileSystem {
	return &filesDataFS{files}
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

// Files implements a http.FileSystem by a list of file name and content file.
func Files(files ...string) http.FileSystem {
	return &filesFS{files}
}

// -----------------------------------------------------------------------------------------
