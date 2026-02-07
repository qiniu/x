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

package filter

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/qiniu/x/http/fs/filter"
	"github.com/qiniu/x/http/fs/ignore"
	"github.com/qiniu/x/http/fsx"
)

// -----------------------------------------------------------------------------------------

const (
	SchemeSelect = "select"
	SchemeIgnore = "ignore"
)

func init() {
	fsx.Register(SchemeSelect, Select)
	fsx.Register(SchemeIgnore, Ignore)
}

// url = `select: <pattern1>;<pattern2>;...;<patternN> | <baseFS>`
func Select(ctx context.Context, url string) (fs http.FileSystem, close fsx.Closer, err error) {
	url = strings.TrimPrefix(url, SchemeSelect+":")
	patterns, baseFS, close, err := parse(ctx, url)
	if err != nil {
		return
	}
	fs = filter.Select(baseFS, patterns...)
	return
}

// url = `ignore: <pattern1>;<pattern2>;...;<patternN> | <baseFS>`
func Ignore(ctx context.Context, url string) (fs http.FileSystem, close fsx.Closer, err error) {
	url = strings.TrimPrefix(url, SchemeIgnore+":")
	patterns, baseFS, close, err := parse(ctx, url)
	if err != nil {
		return
	}
	fs = ignore.New(baseFS, patterns...)
	return
}

func parse(ctx context.Context, url string) (patterns []string, baseFS http.FileSystem, close fsx.Closer, err error) {
	parts := strings.SplitN(url, "|", 2)
	if len(parts) != 2 {
		return nil, nil, nil, fs.ErrInvalid
	}
	patterns = strings.Split(strings.TrimSpace(parts[0]), ";")
	base := strings.TrimLeft(parts[1], " ")
	baseFS, close, err = fsx.Open(ctx, base)
	return
}

// -----------------------------------------------------------------------------------------
