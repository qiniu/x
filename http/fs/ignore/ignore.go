package ignore

import (
	"net/http"

	"github.com/qiniu/x/http/fs/filter"
)

// -----------------------------------------------------------------------------------------

func Matched(ignore []string, fullName, dir, fname string) bool {
	return filter.Matched(ignore, fullName, dir, fname)
}

func New(fs http.FileSystem, patterns ...string) http.FileSystem {
	return filter.New(fs, func(dir string, fi filter.DirEntry) bool {
		return !filter.Matched(patterns, "", dir, fi.Name())
	})
}

// -----------------------------------------------------------------------------------------
