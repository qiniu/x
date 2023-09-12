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

package fsx

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

var (
	ErrUnknownScheme = errors.New("unknown scheme")
)

// -----------------------------------------------------------------------------------------

type Closer = func() error
type Opener = func(ctx context.Context, url string) (http.FileSystem, Closer, error)

var (
	openers = make(map[string]Opener, 4)
)

// Register registers a Opener with specified scheme.
func Register(scheme string, open Opener) {
	openers[scheme] = open
}

// Open opens a file system by the scheme of specified url.
func Open(ctx context.Context, url string) (http.FileSystem, Closer, error) {
	scheme := schemeOf(url)
	if o, ok := openers[scheme]; ok {
		return o(ctx, url)
	}
	return nil, nil, &fs.PathError{Op: "fsx.Open", Err: ErrUnknownScheme, Path: url}
}

func schemeOf(url string) (scheme string) {
	pos := strings.IndexAny(url, ":/")
	if pos > 0 {
		if url[pos] == ':' {
			return url[:pos]
		}
	}
	return ""
}

// -----------------------------------------------------------------------------------------
