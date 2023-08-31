package cached

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Remote interface {
	Init(local string)
	Lstat(localFile string) (fs.FileInfo, error)
	SyncLstat(local string, name string) (fs.FileInfo, error)
	SyncOpen(local string, name string) (http.File, error)
}

type fsCached struct {
	local  string
	remote Remote
}

func (p *fsCached) Open(name string) (f http.File, err error) {
	if !strings.HasPrefix(name, "/") { // name should start with "/"
		name = "/" + name
	}
	remote, local := p.remote, p.local
	localFile := filepath.Join(local, name)
	fi, err := remote.Lstat(localFile)
	if os.IsNotExist(err) {
		fi, err = p.remote.SyncLstat(local, name)
		if err != nil {
			return
		}
	}
	mode := fi.Mode()
	if (mode & fs.ModeSymlink) != 0 {
		return remote.SyncOpen(local, name)
	}
	return os.Open(localFile)
}

func New(local string, remote Remote) http.FileSystem {
	remote.Init(local)
	return &fsCached{local, remote}
}
