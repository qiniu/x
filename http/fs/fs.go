package fs

import (
	"io"
	"io/fs"
	"net/http"
	"os"
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

// Union merge a list of http.FileSystem into a union http.FileSystem object.
func Union(fs ...http.FileSystem) http.FileSystem {
	return &unionFS{fs}
}

// -----------------------------------------------------------------------------------------

type rootDir struct {
}

func (p rootDir) Name() string {
	return "/"
}

func (p rootDir) Size() int64 {
	return 0
}

func (p rootDir) Mode() fs.FileMode {
	return fs.ModeDir
}

func (p rootDir) ModTime() time.Time {
	return time.Now()
}

func (p rootDir) IsDir() bool {
	return true
}

func (p rootDir) Sys() interface{} {
	return nil
}

func (p rootDir) Close() error {
	return nil
}

func (p rootDir) Write(b []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (p rootDir) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

func (p rootDir) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, io.EOF
}

func (p rootDir) Seek(offset int64, whence int) (int64, error) {
	return 0, io.EOF
}

func (p rootDir) Stat() (fs.FileInfo, error) {
	return rootDir{}, nil
}

func (p rootDir) Open(name string) (f http.File, err error) {
	if name == "/" {
		return rootDir{}, nil
	}
	return nil, os.ErrNotExist
}

// Root implents a http.FileSystem that only have a root directory.
func Root() http.FileSystem {
	return rootDir{}
}

// -----------------------------------------------------------------------------------------
