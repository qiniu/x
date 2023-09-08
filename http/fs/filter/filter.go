package filter

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// -----------------------------------------------------------------------------------------

func Matched(patterns []string, fullName, dir, fname string) bool {
	for _, ign := range patterns {
		if strings.HasPrefix(ign, "/") { // start with /
			if fullName == "" {
				fullName = path.Join(dir, fname)
			}
			if strings.HasSuffix(ign, "/") { // end with /
				if strings.HasPrefix(fullName, ign) {
					return true
				}
			} else if strings.HasSuffix(ign, "*") {
				if strings.HasPrefix(fullName, ign[:len(ign)-1]) {
					return true
				}
			} else if fullName == ign {
				return true
			}
		} else {
			if fname == "" {
				dir, fname = path.Split(fullName)
			}
			if strings.HasSuffix(ign, "*") {
				if strings.HasPrefix(fname, ign[:len(ign)-1]) {
					return true
				}
			} else if strings.HasPrefix(ign, "*") {
				if strings.HasSuffix(fname, ign[1:]) {
					return true
				}
			} else if fname == ign {
				return true
			}
		}
	}
	return false
}

// -----------------------------------------------------------------------------------------

type filterDir struct {
	http.File
	dir    string
	filter FilterFunc
}

func (p *filterDir) ReadDir(count int) (fis []fs.DirEntry, err error) {
	f, ok := p.File.(interface {
		ReadDir(count int) ([]fs.DirEntry, error)
	})
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: p.dir, Err: errors.New("not implemented")}
	}
	if fis, err = f.ReadDir(count); err != nil {
		return
	}
	n := 0
	dir, filter := p.dir, p.filter
	for _, fi := range fis {
		if filter(dir, fi) {
			fis[n] = fi
			n++
		}
	}
	return fis[:n], nil
}

func (p *filterDir) Readdir(count int) (fis []fs.FileInfo, err error) {
	if fis, err = p.File.Readdir(count); err != nil {
		return
	}
	n := 0
	dir, filter := p.dir, p.filter
	for _, fi := range fis {
		if filter(dir, fi) {
			fis[n] = fi
			n++
		}
	}
	return fis[:n], nil
}

type fsFilter struct {
	fs     http.FileSystem
	filter FilterFunc
}

func (p *fsFilter) Open(name string) (f http.File, err error) {
	if f, err = p.fs.Open(name); err != nil {
		return
	}
	if fi, err := f.Stat(); err == nil && fi.IsDir() {
		f = &filterDir{f, name, p.filter}
	}
	return
}

// -----------------------------------------------------------------------------------------

type DirEntry interface {
	// Name returns the name of the file (or subdirectory) described by the entry.
	// This name is only the final element of the path (the base name), not the entire path.
	// For example, Name would return "hello.go" not "home/gopher/hello.go".
	Name() string

	// IsDir reports whether the entry describes a directory.
	IsDir() bool
}

type FilterFunc = func(dir string, fi DirEntry) bool

func New(fs http.FileSystem, filter FilterFunc) http.FileSystem {
	return &fsFilter{fs, filter}
}

// -----------------------------------------------------------------------------------------

func Select(fs http.FileSystem, patterns ...string) http.FileSystem {
	return New(fs, func(dir string, fi DirEntry) bool {
		name := fi.Name()
		for _, item := range patterns {
			if strings.HasPrefix(item, "/") && strings.HasSuffix(item, "/") {
				if strings.HasPrefix(item, name) && (name == item || item[len(name)] == '/') {
					return true
				}
			}
		}
		return Matched(patterns, "", dir, name)
	})
}

// -----------------------------------------------------------------------------------------
