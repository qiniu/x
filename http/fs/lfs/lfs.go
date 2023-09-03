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
	if err := cached.DownloadFile(p.localFile, p.file); err != nil {
		log.Println("[WARN] Cache file failed:", err)
	}
	return p.ReadCloser.Close()
}

// -----------------------------------------------------------------------------------------

type fileInfoRemote struct {
	fs.FileInfo
	size int64
}

func (p *fileInfoRemote) Size() int64 {
	return p.size
}

func (p *fileInfoRemote) Mode() fs.FileMode {
	return p.FileInfo.Mode() | cached.ModeRemote
}

func (p *fileInfoRemote) IsDir() bool {
	return false
}

func (p *fileInfoRemote) Sys() interface{} {
	return nil
}

// -----------------------------------------------------------------------------------------

type remote struct {
	exts    map[string]struct{}
	urlBase string
}

func (p *remote) isRemote(fi fs.FileInfo, file string) bool {
	if !fi.Mode().IsRegular() || fi.Size() > 255 {
		return false
	}
	ext := filepath.Ext(file)
	_, ok := p.exts[ext]
	return ok
}

func (p *remote) ReaddirAll(localDir string, dir *os.File, offline bool) (fis []fs.FileInfo, err error) {
	if fis, err = dir.Readdir(-1); err != nil {
		return
	}
	n := 0
	for _, fi := range fis {
		name := fi.Name()
		if p.isRemote(fi, name) {
			if offline {
				continue
			}
			localFile := filepath.Join(localDir, name)
			fi = remoteStat(localFile, fi)
		}
		fis[n] = fi
		n++
	}
	return fis[:n], nil
}

func (p *remote) Lstat(localFile string) (fi fs.FileInfo, err error) {
	fi, err = os.Lstat(localFile)
	if err != nil {
		return
	}
	if p.isRemote(fi, localFile) {
		fi = remoteStat(localFile, fi)
	}
	return
}

const (
	lfsSpec = "version https://git-lfs.github.com/spec/"
	lfsSize = "size "
)

func remoteStat(localFile string, fi fs.FileInfo) fs.FileInfo {
	b, e := os.ReadFile(localFile)
	text := string(b)
	if e != nil || !strings.HasPrefix(text, lfsSpec) {
		return fi
	}
	lines := strings.SplitN(text, "\n", 4)
	for _, line := range lines {
		if strings.HasPrefix(line, lfsSize) {
			if size, e := strconv.ParseInt(line[len(lfsSize):], 10, 64); e == nil {
				return &fileInfoRemote{fi, size}
			}
			break
		}
	}
	return fi
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

func (p *remote) Init(local string, offline bool) {
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

func NewOfflineCached(local string, urlBase string, exts ...string) http.FileSystem {
	return cached.New(local, NewRemote(urlBase, exts...), true)
}

// -----------------------------------------------------------------------------------------
