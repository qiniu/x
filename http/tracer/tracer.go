package tracer

import (
	"log"
	"net/http"
	"time"
)

// -------------------------------------------------------------------------------

type teeResponseWriter struct {
	w []http.ResponseWriter
}

func (p *teeResponseWriter) Header() http.Header {
	return p.w[0].Header()
}

func (p *teeResponseWriter) Write(buf []byte) (n int, err error) {
	n, err = p.w[0].Write(buf)
	for _, w := range p.w[1:] {
		w.Write(buf)
	}
	return
}

func (p *teeResponseWriter) WriteHeader(statusCode int) {
	p.w[0].WriteHeader(statusCode)
	for _, w := range p.w[1:] {
		w.WriteHeader(statusCode)
	}
}

func Tee(w ...http.ResponseWriter) http.ResponseWriter {
	return &teeResponseWriter{w}
}

// -------------------------------------------------------------------------------

type CodeRecorder struct {
	Code int
}

func (p *CodeRecorder) Header() http.Header {
	return nil
}

func (p *CodeRecorder) Write(buf []byte) (n int, err error) {
	return
}

func (p *CodeRecorder) WriteHeader(statusCode int) {
	p.Code = statusCode
}

func NewCodeRecorder() *CodeRecorder {
	return &CodeRecorder{}
}

// -------------------------------------------------------------------------------

func New(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := NewCodeRecorder()
		tee := Tee(w, recorder)
		log.Printf("%s %v\n", r.Method, r.URL)
		start := time.Now()
		h.ServeHTTP(tee, r)
		dur := time.Since(start)
		log.Printf("Returned %d in %d ms\n", recorder.Code, dur.Milliseconds())
	})
}

// -------------------------------------------------------------------------------
