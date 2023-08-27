package nocache

import (
	"bufio"
	"net"
	"net/http"
)

type respWriter struct {
	http.ResponseWriter
}

func (p *respWriter) WriteHeader(statusCode int) {
	w := p.ResponseWriter
	w.Header().Del("Last-Modified")
	w.WriteHeader(statusCode)
}

func New(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&respWriter{w}, r)
	})
}

func (p *respWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return p.ResponseWriter.(http.Hijacker).Hijack()
}
