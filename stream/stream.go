/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stream

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/qiniu/x/byteutil"
)

var (
	// ErrUnknownScheme is returned when an unknown scheme is encountered in a URL.
	ErrUnknownScheme = errors.New("unknown scheme")

	// ErrInvalidSource is returned when an invalid source is encountered.
	ErrInvalidSource = errors.New("invalid source")
)

// -------------------------------------------------------------------------------------

// OpenFunc defines the function type for opening a resource by URL.
type OpenFunc = func(url string) (io.ReadCloser, error)

var (
	openers = map[string]OpenFunc{}
)

// Register registers a scheme with an open function.
func Register(scheme string, open OpenFunc) {
	openers[scheme] = open
}

// Open opens a resource identified by the given URI.
// It supports different schemes by utilizing registered open functions.
// If the URI has no scheme, it is treated as a file path.
func Open(uri string) (io.ReadCloser, error) {
	scheme := schemeOf(uri)
	if scheme == "" {
		return os.Open(uri)
	}
	if open, ok := openers[scheme]; ok {
		return open(uri)
	}
	return nil, &fs.PathError{Op: "stream.Open", Err: ErrUnknownScheme, Path: uri}
}

func schemeOf(uri string) (scheme string) {
	pos := strings.IndexAny(uri, ":/")
	if pos > 0 {
		if uri[pos] == ':' {
			return uri[:pos]
		}
	}
	return ""
}

// -------------------------------------------------------------------------------------

// ReadSource converts src to a []byte if possible; otherwise it returns an error.
// Supported types for src are:
//   - string (as content, NOT as filename)
//   - []byte (as content)
//   - *bytes.Buffer (as content)
//   - io.Reader (as content)
func ReadSource(src any) ([]byte, error) {
	switch s := src.(type) {
	case string:
		return byteutil.Bytes(s), nil
	case []byte:
		return s, nil
	case *bytes.Buffer:
		// is io.Reader, but src is already available in []byte form
		if s != nil {
			return s.Bytes(), nil
		}
	case io.Reader:
		return io.ReadAll(s)
	}
	return nil, ErrInvalidSource
}

// If src != nil, ReadSourceLocal converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, ReadSourceLocal returns
// the result of reading the file specified by filename.
func ReadSourceLocal(filename string, src any) ([]byte, error) {
	if src != nil {
		return ReadSource(src)
	}
	return os.ReadFile(filename)
}

// ReadSourceFromURI reads the source from the given URI.
// If src != nil, it reads from src; otherwise, it opens the URI and reads
// from it.
func ReadSourceFromURI(uri string, src any) (ret []byte, err error) {
	if src == nil {
		var f io.ReadCloser
		f, err = Open(uri)
		if err != nil {
			return
		}
		defer f.Close()
		src = f
	}
	return ReadSource(src)
}

// -------------------------------------------------------------------------------------
