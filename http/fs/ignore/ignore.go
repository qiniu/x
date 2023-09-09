package ignore

import (
	"net/http"

	"github.com/qiniu/x/http/fs/filter"
)

// -----------------------------------------------------------------------------------------

func Matched(ignore []string, fullName, dir, fname string, isDir bool) bool {
	return filter.Matched(ignore, fullName, dir, fname, isDir)
}

// New creates a http.FileSystem with ignore patterns.
func New(fs http.FileSystem, patterns ...string) http.FileSystem {
	return filter.New(fs, func(name string, fi filter.DirEntry) bool {
		return !filter.Matched(patterns, name, "", "", fi.IsDir())
	})
}

// -----------------------------------------------------------------------------------------
