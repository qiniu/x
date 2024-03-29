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
