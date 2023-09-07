package fs

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// -----------------------------------------------------------------------------------------

type HttpOpener struct {
	Client *http.Client
	Header http.Header
}

// Open opens a http.File from an url.
func (p *HttpOpener) Open(ctx context.Context, url string) (file http.File, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	if h := p.Header; h != nil {
		req.Header = h
	}

	c := p.Client
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		e := &fs.PathError{Op: "http.Get", Path: url}
		if resp.StatusCode == 404 {
			e.Err = fs.ErrNotExist
			return nil, e
		}
		url := "url"
		if req := resp.Request; req != nil {
			url = req.URL.String()
		}
		e.Err = fmt.Errorf("http.Get %s error: status %d (%s)", url, resp.StatusCode, resp.Status)
		return nil, e
	}
	return HttpFile(resp.Request.URL.Path, resp), nil
}

// -----------------------------------------------------------------------------------------

type HttpFS struct {
	HttpOpener
	urlBase string
	ctx     context.Context
}

// Open is required by http.File.
func (p *HttpFS) Open(name string) (file http.File, err error) {
	return p.HttpOpener.Open(p.ctx, p.urlBase+name)
}

// With specifies http.Client and http.Header used by http.Get.
func (fs HttpFS) With(client *http.Client, header http.Header) *HttpFS {
	fs.Client, fs.Header = client, header
	return &fs
}

// Http creates a HttpFS which implements a http.FileSystem by http.Get join(urlBase, name).
func Http(urlBase string, ctx ...context.Context) *HttpFS {
	fs := &HttpFS{urlBase: strings.TrimSuffix(urlBase, "/")}
	if ctx != nil {
		fs.ctx = ctx[0]
	} else {
		fs.ctx = context.Background()
	}
	return fs
}

// -----------------------------------------------------------------------------------------

type fsWithTracker struct {
	fs     http.FileSystem
	exts   map[string]struct{}
	httpfs *HttpFS
}

func (p *fsWithTracker) Open(name string) (file http.File, err error) {
	ext := path.Ext(name)
	if _, ok := p.exts[ext]; !ok {
		return p.fs.Open(name)
	}
	return p.httpfs.Open(name)
}

// WithTracker implements a http.FileSystem by pactching large file access like git lfs.
// Here trackerInit should be (urlBase string) or (httpfs *fs.HttpFS).
func WithTracker(fs http.FileSystem, trackerInit interface{}, exts ...string) http.FileSystem {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	var httpfs *HttpFS
	switch tracker := trackerInit.(type) {
	case string: // urlBase string
		httpfs = Http(tracker)
	case *HttpFS:
		httpfs = tracker
	default:
		panic("fs.WithTracker: trackerInit should be (urlBase string) or (httpfs *fs.HttpFS)")
	}
	return &fsWithTracker{fs, m, httpfs}
}

// -----------------------------------------------------------------------------------------
