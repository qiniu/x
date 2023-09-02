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
	"time"

	xfs "github.com/qiniu/x/http/fs"
	"github.com/qiniu/x/http/fs/cached"
)

var (
	CacheFileName string = ".cache"
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

type dirEntry struct {
	Name    string      `msg:"name"`  // base name of the file
	Size    int64       `msg:"size"`  // length in bytes for regular files; system-dependent for others
	ModTime int64       `msg:"mtime"` // modification time in UnixMicro
	Mode    fs.FileMode `msg:"mode"`  // file mode bits
}

type fileInfo struct {
	d dirEntry
}

func (p *fileInfo) Name() string {
	return p.d.Name
}

func (p *fileInfo) Size() int64 {
	return p.d.Size
}

func (p *fileInfo) Mode() fs.FileMode {
	return p.d.Mode
}

func (p *fileInfo) ModTime() time.Time {
	return time.UnixMicro(p.d.ModTime)
}

func (p *fileInfo) IsDir() bool {
	return p.d.Mode.IsDir()
}

func (p *fileInfo) Sys() interface{} {
	return nil
}

// -----------------------------------------------------------------------------------------

type remote struct {
	exts    map[string]struct{}
	urlBase string
}

func (p *remote) notIsRemote(fi fs.FileInfo, file string) bool {
	if !fi.Mode().IsRegular() || fi.Size() > 255 {
		return true
	}
	ext := filepath.Ext(file)
	_, ok := p.exts[ext]
	return !ok
}

func (p *remote) ReaddirAll(localDir string, dir *os.File) (fis []fs.FileInfo, err error) {
	// cacheFile := filepath.Join(localDir, CacheFileName)
	return
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
	if p.notIsRemote(fi, localFile) {
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
				return &fileInfoRemote{fi, size}, nil
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
