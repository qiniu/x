package plugins

import (
	"net/http"
	"path"
)

type handler = func(w http.ResponseWriter, r *http.Request, next http.Handler)

func New(h http.Handler, plugins ...interface{}) http.Handler {
	n := len(plugins)
	exts := make(map[string]handler, n/2)
	for i := 0; i < n; i += 2 {
		ext := plugins[i].(string)
		fn := plugins[i+1].(handler)
		exts[ext] = fn
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := path.Ext(r.URL.Path)
		if fn, ok := exts[ext]; ok {
			fn(w, r, h)
			return
		}
		h.ServeHTTP(w, r)
	})
}
