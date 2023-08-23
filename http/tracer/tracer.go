package tracer

import (
	"log"
	"net/http"
)

func New(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %v\n", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
