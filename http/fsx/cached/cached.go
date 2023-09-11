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

type config struct {
	Base      string `json:"base"` // url of base file system
	CacheFile bool   `json:"cacheFile"`
}

func New(ctx context.Context, localDir string, offline ...bool) (fs http.FileSystem, close fsx.Closer, err error) {
	confFile := filepath.Join(localDir, confFileName)
	b, err := os.ReadFile(confFile)
	if err != nil {
		return
	}
	var conf config
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

// -----------------------------------------------------------------------------------------
