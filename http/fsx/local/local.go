package local

import (
	"context"
	"net/http"

	"github.com/qiniu/x/http/fsx"
)

const (
	Scheme = ""
)

func init() {
	fsx.Register(Scheme, Open)
}

// Open opens a local file system.
func Open(ctx context.Context, url string) (http.FileSystem, fsx.Closer, error) {
	return http.Dir(url), nil, nil
}

// Check checks a file system is local or not.
func Check(fsys http.FileSystem) (string, bool) {
	d, ok := fsys.(http.Dir)
	return string(d), ok
}

// -----------------------------------------------------------------------------------------
