package fs

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
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

type fsPlugins struct {
	fs   http.FileSystem
	exts map[string]Opener
}

func (p *fsPlugins) Open(name string) (http.File, error) {
	ext := path.Ext(name)
	if fn, ok := p.exts[ext]; ok {
		return fn(p.fs, name)
	}
	return p.fs.Open(name)
}

type Opener = func(fs http.FileSystem, name string) (file http.File, err error)

// Plugins implements a filesystem with plugins by specified (ext string, plugin Opener) pairs.
func Plugins(fs http.FileSystem, plugins ...interface{}) http.FileSystem {
	n := len(plugins)
	exts := make(map[string]Opener, n/2)
	for i := 0; i < n; i += 2 {
		ext := plugins[i].(string)
		fn := plugins[i+1].(Opener)
		exts[ext] = fn
	}
	return &fsPlugins{fs, exts}
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

// Root implements a http.FileSystem that only have a root directory.
func Root() http.FileSystem {
	return rootDir{}
}

// -----------------------------------------------------------------------------------------
