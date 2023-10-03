/*
 Copyright 2023 Qiniu Limited (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package fallback

import (
	"net/http"
)

// Fallback is a http.Handler that fallback to second handler if status code of
// first handler is in fallbackStatus
func New(fallbackStatus []int, first http.Handler, second http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		first.ServeHTTP(newFallbackWriter(fallbackStatus, w, r, second), r)
	})
}

type fallbackWriter struct {
	fallbackStatus []int
	header         http.Header
	writer         http.ResponseWriter
	reader         *http.Request
	nextHandler    http.Handler
	open           bool
}

func newFallbackWriter(fallbackStatus []int, w http.ResponseWriter, r *http.Request, h http.Handler) http.ResponseWriter {
	return &fallbackWriter{
		fallbackStatus: fallbackStatus,
		writer:         w,
		reader:         r,
		nextHandler:    h,
		header:         make(http.Header),
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (p *fallbackWriter) WriteHeader(code int) {
	if contains(p.fallbackStatus, code) {
		p.nextHandler.ServeHTTP(p.writer, p.reader)
	} else {
		p.open = true
		// Don't write header to writer before status code is determined
		for k, v := range p.header {
			p.writer.Header()[k] = v
		}
		p.writer.WriteHeader(code)
	}
}

func (p *fallbackWriter) Header() http.Header {
	if p.open {
		return p.writer.Header()
	}
	return p.header
}

func (p *fallbackWriter) Write(data []byte) (int, error) {
	if p.open {
		return p.writer.Write(data)
	}
	return len(data), nil
}
