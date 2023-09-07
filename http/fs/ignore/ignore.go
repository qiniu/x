package ignore

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// -----------------------------------------------------------------------------------------

func Matched(ignore []string, fullName, dir, fname string) bool {
	for _, ign := range ignore {
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

type ignDir struct {
	http.File
	dir    string
	ignore []string
}

func (p *ignDir) ReadDir(count int) (fis []fs.DirEntry, err error) {
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
	dir, ignore := p.dir, p.ignore
	for _, fi := range fis {
		if Matched(ignore, "", dir, fi.Name()) {
			continue
		}
		fis[n] = fi
		n++
	}
	return fis[:n], nil
}

func (p *ignDir) Readdir(count int) (fis []fs.FileInfo, err error) {
	if fis, err = p.File.Readdir(count); err != nil {
		return
	}
	n := 0
	dir, ignore := p.dir, p.ignore
	for _, fi := range fis {
		if Matched(ignore, "", dir, fi.Name()) {
			continue
		}
		fis[n] = fi
		n++
	}
	return fis[:n], nil
}

type fsIgnore struct {
	fs     http.FileSystem
	ignore []string
}

func (p *fsIgnore) Open(name string) (f http.File, err error) {
	if f, err = p.fs.Open(name); err != nil {
		return
	}
	if fi, err := f.Stat(); err == nil && fi.IsDir() {
		f = &ignDir{f, name, p.ignore}
	}
	return
}

func New(fs http.FileSystem, patterns ...string) http.FileSystem {
	return &fsIgnore{fs, patterns}
}

// -----------------------------------------------------------------------------------------
