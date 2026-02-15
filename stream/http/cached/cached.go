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

package cached

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"

	"github.com/qiniu/x/stream"
	"github.com/qiniu/x/stream/http"
)

// -------------------------------------------------------------------------------------

var (
	cacheDir string
	errInit  error
)

func init() {
	root, err := os.UserCacheDir()
	if err != nil {
		errInit = err
		return
	}
	cacheDir = root + "/qiniu.x.http/"
	errInit = os.MkdirAll(cacheDir, 0755)
}

// -------------------------------------------------------------------------------------

// writeCache writes the http response to cache file.
func writeCache(cacheDir, fname string, url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	f, err := os.CreateTemp(cacheDir, "*.cache~")
	if err != nil {
		return
	}
	tempFile := f.Name()
	_, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tempFile)
	} else if err = os.Rename(tempFile, cacheDir+fname); err != nil {
		os.Remove(tempFile)
	}
	return
}

// readCache reads the cache file and returns a ReadCloser.
func readCache(cacheFile string, _ fs.FileInfo) (ret io.ReadCloser, err error) {
	return os.Open(cacheFile)
}

// Open opens a http file object.
func Open(url_ string) (ret io.ReadCloser, err error) {
	if errInit != nil {
		return http.Open(url_) // fallback to direct open
	}
	u, err := url.Parse(url_)
	if err != nil {
		return
	}
	fname := path.Base(u.Path)
	ext := path.Ext(fname)
	hash := md5.Sum([]byte(url_))
	hashstr := base64.RawURLEncoding.EncodeToString(hash[:])
	fname = fmt.Sprintf("%s-%s%s", fname[:len(fname)-len(ext)], hashstr, ext)
	file := cacheDir + fname
	if fi, e := os.Stat(file); e == nil {
		if ret, err = readCache(file, fi); err == nil { // cache hit
			return
		}
	}
	if err = writeCache(cacheDir, fname, url_); err != nil {
		return // write cache failed
	}
	return readCache(file, nil)
}

func init() {
	stream.Register("http", Open)
	stream.Register("https", Open)
}

// -------------------------------------------------------------------------------------
