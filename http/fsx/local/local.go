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

package local

import (
	"context"
	"net/http"

	"github.com/qiniu/x/http/fsx"
)

const (
	Scheme = ""
)

func init() {
	fsx.Register(Scheme, Open)
}

// Open opens a local file system.
func Open(ctx context.Context, url string) (http.FileSystem, fsx.Closer, error) {
	return http.Dir(url), nil, nil
}

// Check checks a file system is local or not.
func Check(fsys http.FileSystem) (string, bool) {
	d, ok := fsys.(http.Dir)
	return string(d), ok
}

// -----------------------------------------------------------------------------------------
