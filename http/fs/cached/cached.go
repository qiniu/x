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
		file = xfs.Dir(f, fis)
	} else {
		file = f
	}
	return
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
