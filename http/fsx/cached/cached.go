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

package cached

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiniu/x/http/fs/cached/remote"
	"github.com/qiniu/x/http/fsx"
)

const (
	Scheme = "cached"
)

func init() {
	fsx.Register(Scheme, Open)
}

// url = "cached:<localDir>"
func Open(ctx context.Context, url string) (fs http.FileSystem, _ fsx.Closer, err error) {
	url = strings.TrimPrefix(url, Scheme+":")
	return New(ctx, url)
}

const (
	confFileName = remote.SysFilePrefix + "conf"
)

type Config struct {
	Base      string `json:"base"` // url of base file system
	CacheFile bool   `json:"cacheFile"`
}

func New(ctx context.Context, localDir string, offline ...bool) (fs http.FileSystem, close fsx.Closer, err error) {
	confFile := filepath.Join(localDir, confFileName)
	b, err := os.ReadFile(confFile)
	if err != nil {
		return
	}
	var conf Config
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return
	}
	base, close, err := fsx.Open(ctx, conf.Base)
	if err != nil {
		return
	}
	fs, err = remote.NewCached(localDir, base, nil, conf.CacheFile, offline...)
	if err != nil && close != nil {
		close()
		close = nil
	}
	return
}

func WriteConf(localDir string, conf *Config) (err error) {
	confFile := filepath.Join(localDir, confFileName)
	b, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return
	}
	return os.WriteFile(confFile, b, 0644)
}

// -----------------------------------------------------------------------------------------
