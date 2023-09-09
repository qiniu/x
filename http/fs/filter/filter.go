package filter

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// -----------------------------------------------------------------------------------------

func Matched(patterns []string, fullName, dir, fname string, isDir bool) bool {
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
			} else if ign == fullName {
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
				if !isDir {
					// the pattern `*.` means to match filename without extension
					if strings.HasSuffix(ign, ".") {
						if path.Ext(fname) != "" {
							continue
						}
						ign = ign[:len(ign)-1]
					}
				}
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

func Selected(patterns []string, name string, isDir bool) bool {
	if isDir {
		return selectDir(patterns, name)
	}
	return Matched(patterns, name, "", "", false)
}

func selectDir(patterns []string, fullName string) bool {
	for _, ign := range patterns {
		if strings.HasPrefix(ign, "/") { // start with /
			if strings.HasSuffix(ign, "/") { // end with /
				if strings.HasPrefix(fullName, ign) {
					return true
				}
				if strings.HasPrefix(ign, fullName) && (fullName == ign || ign[len(fullName)] == '/') {
					return true
				}
			} else if strings.HasSuffix(ign, "*") {
				prefix := ign[:len(ign)-1]
				if strings.HasPrefix(fullName, prefix) {
					return true
				}
				if strings.HasPrefix(prefix, fullName) {
					return true
				}
			} else if ign == fullName {
				return true
			} else if strings.HasPrefix(ign, fullName) && ign[len(fullName)] == '/' {
				return true
			}
		} else {
			return true
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
		if filter(dir+fi.Name(), fi) {
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
		if filter(dir+fi.Name(), fi) {
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
	if name == "/" { // don't filter root directory
		return &filterDir{f, "/", p.filter}, nil
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return
	}
	if !p.filter(name, fi) {
		f.Close()
		return nil, fs.ErrNotExist
	}
	if fi.IsDir() {
		f = &filterDir{f, name + "/", p.filter}
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

type FilterFunc = func(name string, fi DirEntry) bool

// New creates a http.FileSystem with a (filter FilterFunc).
func New(fs http.FileSystem, filter FilterFunc) http.FileSystem {
	return &fsFilter{fs, filter}
}

// -----------------------------------------------------------------------------------------

// Select creates a http.FileSystem with filter patterns.
func Select(fs http.FileSystem, patterns ...string) http.FileSystem {
	return New(fs, func(name string, fi DirEntry) bool {
		return Selected(patterns, name, fi.IsDir())
	})
}

// -----------------------------------------------------------------------------------------
