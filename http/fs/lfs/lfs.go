package lfs

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	xfs "github.com/qiniu/x/http/fs"
	"github.com/qiniu/x/http/fs/cached"
)

// -----------------------------------------------------------------------------------------

type cachedCloser struct {
	io.ReadCloser
	file      http.File
	localFile string
}

func (p *cachedCloser) Close() error {
	file := p.file
	_, err := file.Seek(0, io.SeekStart)
	if err == nil {
		localFile := p.localFile
		localFileDownloading := localFile + ".download" // TODO: use tempfile
		err = xfs.Download(localFileDownloading, file)
		if err == nil {
			err = os.Rename(localFileDownloading, localFile)
			if err != nil {
				log.Println("Cache failed:", err)
			}
		}
	}
	return p.ReadCloser.Close()
}

// -----------------------------------------------------------------------------------------

type fileInfo struct {
	fs.FileInfo
	size int64
}

func (p *fileInfo) Size() int64 {
	return p.size
}

func (p *fileInfo) Mode() fs.FileMode {
	return p.FileInfo.Mode() | fs.ModeSymlink
}

func (p *fileInfo) IsDir() bool {
	return false
}

func (p *fileInfo) Sys() interface{} {
	return nil
}

type remote struct {
	exts    map[string]struct{}
	urlBase string
}

const (
	lfsSpec = "version https://git-lfs.github.com/spec/"
	lfsSize = "size "
)

func (p *remote) Lstat(localFile string) (fi fs.FileInfo, err error) {
	fi, err = os.Lstat(localFile)
	if err != nil {
		return
	}
	mode := fi.Mode()
	if !mode.IsRegular() || fi.Size() > 255 {
		return
	}
	ext := filepath.Ext(localFile)
	if _, ok := p.exts[ext]; !ok {
		return
	}
	b, e := os.ReadFile(localFile)
	text := string(b)
	if e != nil || !strings.HasPrefix(text, lfsSpec) {
		return
	}

	lines := strings.SplitN(text, "\n", 4)
	for _, line := range lines {
		if strings.HasPrefix(line, lfsSize) {
			if size, e := strconv.ParseInt(line[len(lfsSize):], 10, 64); e == nil {
				return &fileInfo{fi, size}, nil
			}
			break
		}
	}
	return
}

func (p *remote) SyncLstat(local string, name string) (fs.FileInfo, error) {
	return nil, os.ErrNotExist
}

func (p *remote) SyncOpen(local string, name string) (f http.File, err error) {
	resp, err := http.Get(p.urlBase + name)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		url := "url"
		if req := resp.Request; req != nil {
			url = req.URL.String()
		}
		return nil, fmt.Errorf("http.Get %s error: status %d (%s)", url, resp.StatusCode, resp.Status)
	}
	localFile := filepath.Join(local, name)
	closer := &cachedCloser{resp.Body, nil, localFile}
	resp.Body = closer
	closer.file = xfs.HttpFile(name, resp)
	return closer.file, nil
}

func (p *remote) Init(local string) {
}

func NewRemote(urlBase string, exts ...string) cached.Remote {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return &remote{m, strings.TrimSuffix(urlBase, "/")}
}

func NewCached(local string, urlBase string, exts ...string) http.FileSystem {
	return cached.New(local, NewRemote(urlBase, exts...))
}

// -----------------------------------------------------------------------------------------
