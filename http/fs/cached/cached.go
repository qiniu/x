package cached

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	xfs "github.com/qiniu/x/http/fs"
)

var (
	ErrOffline = errors.New("remote filesystem is offline")
)

const (
	ModeRemote = fs.ModeSymlink | fs.ModeIrregular
)

func IsRemote(mode fs.FileMode) bool {
	return (mode & ModeRemote) == ModeRemote
}

// -----------------------------------------------------------------------------------------

type Remote interface {
	Init(local string, offline bool)
	Lstat(localFile string) (fs.FileInfo, error)
	ReaddirAll(localDir string, dir *os.File, offline bool) (fis []fs.FileInfo, err error)
	SyncLstat(local string, name string) (fs.FileInfo, error)
	SyncOpen(local string, name string) (http.File, error)
}

type fsCached struct {
	local   string
	remote  Remote
	offline bool
}

func New(local string, remote Remote, offline ...bool) http.FileSystem {
	var isOffline bool
	if offline != nil {
		isOffline = offline[0]
	}
	remote.Init(local, isOffline)
	return &fsCached{local, remote, isOffline}
}

func RemoteOf(fs http.FileSystem) (r Remote, ok bool) {
	c, ok := fs.(*fsCached)
	if ok {
		r = c.remote
	}
	return
}

func IsOffline(fs http.FileSystem) bool {
	if c, ok := fs.(*fsCached); ok {
		return c.offline
	}
	return false
}

func (p *fsCached) Open(name string) (file http.File, err error) {
	remote, local := p.remote, p.local
	localFile := filepath.Join(local, name)
	fi, err := remote.Lstat(localFile)
	if os.IsNotExist(err) {
		if p.offline {
			return nil, ErrOffline
		}
		fi, err = p.remote.SyncLstat(local, name)
		if err != nil {
			return
		}
	}
	if IsRemote(fi.Mode()) {
		if p.offline {
			return nil, ErrOffline
		}
		return remote.SyncOpen(local, name)
	}
	f, err := os.Open(localFile)
	if err != nil {
		err.(*os.PathError).Path = name
		return
	}
	if fi.IsDir() {
		fis, e := remote.ReaddirAll(localFile, f, p.offline)
		if e != nil {
			f.Close()
			return nil, e
		}
		file = Dir(f, fis)
	} else {
		file = f
	}
	return
}

// -----------------------------------------------------------------------------------------

type StatCloser interface {
	Stat() (fs.FileInfo, error)
	Close() error
}

type dir struct {
	items []fs.FileInfo
	file  StatCloser
	off   int
}

func Dir(base StatCloser, fis []fs.FileInfo) http.File {
	return &dir{fis, base, 0}
}

func (p *dir) Close() error {
	return p.file.Close()
}

func (p *dir) Read(b []byte) (n int, err error) {
	return 0, fs.ErrPermission
}

func (p *dir) Seek(offset int64, whence int) (int64, error) {
	return 0, fs.ErrPermission
}

func (p *dir) Stat() (fs.FileInfo, error) {
	return p.file.Stat()
}

func (p *dir) Readdir(n int) (fis []fs.FileInfo, err error) {
	fis = p.items[p.off:]
	if n <= 0 {
		p.off = len(p.items)
		return
	}
	if len(fis) > n {
		fis = fis[:n]
	} else {
		err = io.EOF
	}
	p.off += len(fis)
	return
}

func (p *dir) ReadDir(n int) (items []fs.DirEntry, err error) {
	fis, err := p.Readdir(n)
	if err != nil && err != io.EOF {
		return
	}
	items = make([]fs.DirEntry, len(fis))
	for i, fi := range fis {
		items[i] = dirEntry{fi}
	}
	return
}

type dirEntry struct {
	fs.FileInfo
}

func (d dirEntry) Info() (fs.FileInfo, error) {
	return d.FileInfo, nil
}

func (d dirEntry) Type() fs.FileMode {
	return d.FileInfo.Mode().Type()
}

// -----------------------------------------------------------------------------------------

func DownloadFile(localFile string, file http.File) (err error) {
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	localFileDownloading := localFile + ".download~"
	err = xfs.Download(localFileDownloading, file)
	if err == nil {
		err = os.Rename(localFileDownloading, localFile)
	}
	if err != nil {
		os.Remove(localFileDownloading)
	}
	return
}

// -----------------------------------------------------------------------------------------
