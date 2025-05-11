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
	// ErrOffline indicates that the remote filesystem is offline.
	ErrOffline = errors.New("remote filesystem is offline")
)

const (
	// ModeRemote indicates that the file is a remote file or directory.
	ModeRemote = fs.ModeSymlink | fs.ModeIrregular
)

// IsRemote checks if the file mode indicates a remote file or directory.
func IsRemote(mode fs.FileMode) bool {
	return (mode & ModeRemote) == ModeRemote
}

// -----------------------------------------------------------------------------------------

// Remote is an interface for remote file system operations.
type Remote interface {
	// Init initializes the remote file system with it local cache.
	Init(local string, offline bool) error

	// Lstat retrieves the cached FileInfo for the specified file or directory.
	Lstat(localFile string) (fs.FileInfo, error)

	// ReaddirAll reads all entries in the directory and returns their cached FileInfo.
	ReaddirAll(localDir string, dir *os.File, offline bool) (fis []fs.FileInfo, err error)

	// SyncLstat retrieves the FileInfo from the remote.
	SyncLstat(local string, name string) (fs.FileInfo, error)

	// SyncOpen retrieves the file from the remote.
	SyncOpen(local string, name string, fi fs.FileInfo) (http.File, error)
}

type fsCached struct {
	local   string
	remote  Remote
	offline bool
}

// New creates a new cached file system with the specified local cache directory
// and remote file system.
func New(local string, remote Remote, offline ...bool) http.FileSystem {
	fs, err := NewEx(local, remote, offline...)
	if err != nil {
		panic(err)
	}
	return fs
}

// NewEx creates a new cached file system with the specified local cache directory
// and remote file system.
func NewEx(local string, remote Remote, offline ...bool) (_ http.FileSystem, err error) {
	var isOffline bool
	if offline != nil {
		isOffline = offline[0]
	}
	err = remote.Init(local, isOffline)
	if err != nil {
		return
	}
	return &fsCached{local, remote, isOffline}, nil
}

// RemoteOf retrieves the remote file system from the cached file system.
func RemoteOf(fs http.FileSystem) (r Remote, ok bool) {
	c, ok := fs.(*fsCached)
	if ok {
		r = c.remote
	}
	return
}

// IsOffline checks if the cached file system is in offline mode.
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
		return remote.SyncOpen(local, name, fi)
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

// DownloadFile downloads the file from the remote to the local cache file.
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
