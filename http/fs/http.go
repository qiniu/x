package fs

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

// -----------------------------------------------------------------------------------------

// HttpOpen opens a http.File from an url.
func HttpOpen(url string) (file http.File, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		url := "url"
		if req := resp.Request; req != nil {
			url = req.URL.String()
		}
		return nil, fmt.Errorf("http.Get %s error: status %d (%s)", url, resp.StatusCode, resp.Status)
	}
	return HttpFile(resp.Request.URL.Path, resp), nil
}

// -----------------------------------------------------------------------------------------

type fsHttp struct {
	urlBase string
}

func (r fsHttp) Open(name string) (file http.File, err error) {
	return HttpOpen(r.urlBase + name)
}

// Http implements a http.FileSystem by http.Get join(urlBase, name).
func Http(urlBase string) http.FileSystem {
	return fsHttp{strings.TrimSuffix(urlBase, "/")}
}

// -----------------------------------------------------------------------------------------

type fsWithTracker struct {
	fs      http.FileSystem
	exts    map[string]struct{}
	urlBase string
}

func (p *fsWithTracker) Open(name string) (file http.File, err error) {
	ext := path.Ext(name)
	if _, ok := p.exts[ext]; !ok {
		return p.fs.Open(name)
	}
	return HttpOpen(p.urlBase + name)
}

// WithTracker implements a http.FileSystem by pactching large file access like git lfs.
func WithTracker(fs http.FileSystem, urlBase string, exts ...string) http.FileSystem {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return &fsWithTracker{fs, m, strings.TrimSuffix(urlBase, "/")}
}

// -----------------------------------------------------------------------------------------
