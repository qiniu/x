/*
 Copyright 2020 Qiniu Limited (qiniu.com)

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

package mockhttp

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/qiniu/x/log"
)

var (
	ErrServerNotFound = errors.New("server not found")
)

// --------------------------------------------------------------------

type mockServerRequestBody struct {
	reader      io.Reader
	closeSignal bool
}

func (r *mockServerRequestBody) Read(p []byte) (int, error) {
	if r.closeSignal || r.reader == nil {
		return 0, io.EOF
	}
	return r.reader.Read(p)
}

func (r *mockServerRequestBody) Close() error {
	r.closeSignal = true
	if c, ok := r.reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// --------------------------------------------------------------------

// Transport is an mock implementation of RoundTripper that supports HTTP,
// HTTPS, and HTTP proxies (for either HTTP or HTTPS with CONNECT).
type Transport struct {
	route      map[string]http.Handler
	remoteAddr string
}

// NewTransport creates a new mock RoundTripper object.
func NewTransport() *Transport {
	return &Transport{
		route:      make(map[string]http.Handler),
		remoteAddr: "127.0.0.1:13579",
	}
}

// SetRemoteAddr sets the remote network address.
func (p *Transport) SetRemoteAddr(remoteAddr string) *Transport {
	p.remoteAddr = remoteAddr
	return p
}

// ListenAndServe listens on a mock network address addr and handler
// to handle requests on incoming connections.
func (p *Transport) ListenAndServe(host string, h http.Handler) error {
	if h == nil {
		h = http.DefaultServeMux
	}
	p.route[host] = h
	return nil
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (p *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	h := p.route[req.URL.Host]
	if h == nil {
		log.Warn("Server not found:", req.Host, "-", req.URL.Host)
		return nil, ErrServerNotFound
	}

	cp := *req
	cp.RemoteAddr = p.remoteAddr
	cp.Body = &mockServerRequestBody{req.Body, false}
	req = &cp

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	req.Body.Close()

	ctlen := int64(-1)
	if v := rw.Header().Get("Content-Length"); v != "" {
		ctlen, _ = strconv.ParseInt(v, 10, 64)
	}

	return &http.Response{
		Status:           "",
		StatusCode:       rw.Code,
		Header:           rw.Header(),
		Body:             io.NopCloser(rw.Body),
		ContentLength:    ctlen,
		TransferEncoding: nil,
		Close:            false,
		Trailer:          nil,
		Request:          req,
	}, nil
}

// --------------------------------------------------------------------

var DefaultTransport = NewTransport()
var DefaultClient = &http.Client{Transport: DefaultTransport}

// ListenAndServe uses DefaultTransport to listen on a mock network address
// addr and handler to handle requests on incoming connections.
func ListenAndServe(host string, h http.Handler) error {
	return DefaultTransport.ListenAndServe(host, h)
}

// --------------------------------------------------------------------
