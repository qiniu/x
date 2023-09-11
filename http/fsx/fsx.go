package fsx

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

var (
	ErrUnknownScheme = errors.New("unknown scheme")
)

// -----------------------------------------------------------------------------------------

type Closer = func() error
type Opener = func(ctx context.Context, url string) (http.FileSystem, Closer, error)

var (
	openers = make(map[string]Opener, 4)
)

// Register registers a Opener with specified scheme.
func Register(scheme string, open Opener) {
	openers[scheme] = open
}

// Open opens a file system by the scheme of specified url.
func Open(ctx context.Context, url string) (http.FileSystem, Closer, error) {
	scheme := schemeOf(url)
	if o, ok := openers[scheme]; ok {
		return o(ctx, url)
	}
	return nil, nil, &fs.PathError{Op: "fsx.Open", Err: ErrUnknownScheme, Path: url}
}

func schemeOf(url string) (scheme string) {
	pos := strings.IndexAny(url, ":/")
	if pos > 0 {
		if url[pos] == ':' {
			return url[:pos]
		}
	}
	return ""
}

// -----------------------------------------------------------------------------------------
