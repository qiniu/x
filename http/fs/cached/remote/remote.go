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

package remote

import (
	"context"
	"encoding/base64"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/x/http/fs/cached"
	xdir "github.com/qiniu/x/http/fs/cached/dir"
)

var (
	debugNet bool
)

const (
	DbgFlagNetwork = 1 << iota
	DbgFlagAll     = DbgFlagNetwork
)

func SetDebug(dbgFlags int) {
	debugNet = (dbgFlags & DbgFlagNetwork) != 0
}

// -----------------------------------------------------------------------------------------

const (
	SysFilePrefix    = ".fscache."
	dirListCacheFile = SysFilePrefix + "ls"
)

func checkDirCached(dir string) fs.FileInfo {
	cacheFile := filepath.Join(dir, dirListCacheFile)
	fi, err := os.Lstat(cacheFile)
	if err != nil {
		fi = nil
	}
	return fi
}

func TouchDirCached(dir string) error {
	cacheFile := filepath.Join(dir, dirListCacheFile)
	return os.WriteFile(cacheFile, nil, 0666)
}

func WriteStubFile(localFile string, fi fs.FileInfo) error {
	if fi.IsDir() {
		return os.Mkdir(localFile, 0755)
	}
	fi = &fileInfoRemote{fi}
	b := xdir.BytesFileInfo(fi)
	dest := base64.URLEncoding.EncodeToString(b)
	// don't need to restore mtime for symlink (saved by BytesFileInfo)
	return os.Symlink(dest, localFile)
}

func readStubFile(localFile string, fi fs.FileInfo) fs.FileInfo {
	dest, e1 := os.Readlink(localFile)
	if e1 == nil {
		if b, e2 := base64.URLEncoding.DecodeString(dest); e2 == nil {
			if ret, e3 := xdir.FileInfoFrom(b); e3 == nil {
				return ret
			}
		}
	}
	return fi
}

func isRemote(fi fs.FileInfo) bool {
	return (fi.Mode() & fs.ModeSymlink) != 0
}

// -----------------------------------------------------------------------------------------

type readdirFile interface {
	Readdir(n int) ([]fs.FileInfo, error)
}

func readdir(f http.File) ([]fs.FileInfo, error) {
	if r, ok := f.(readdirFile); ok {
		return r.Readdir(-1)
	}
	log.Panicln("[FATAL] Readdir notimpl")
	return nil, fs.ErrInvalid
}

// -----------------------------------------------------------------------------------------

type objFile struct {
	http.File
	localFile string
	notify    NotifyFile
}

func (p *objFile) Close() error {
	file := p.File
	if err := cached.DownloadFile(p.localFile, file); err == nil {
		fi, _ := file.Stat()
		os.Chtimes(p.localFile, time.Time{}, fi.ModTime()) // restore mtime
		if notify := p.notify; notify != nil {
			name := file.(interface{ FullName() string }).FullName()
			notify.NotifyFile(context.Background(), name, fi)
		}
	} else {
		log.Println("[WARN] Cache file failed:", err)
	}
	return file.Close()
}

type fileInfoRemote struct {
	fs.FileInfo
}

func (p *fileInfoRemote) Mode() fs.FileMode {
	return p.FileInfo.Mode() | cached.ModeRemote
}

// -----------------------------------------------------------------------------------------

type remote struct {
	bucket    http.FileSystem
	notify    NotifyFile
	cacheFile bool
}

func (p *remote) ReaddirAll(localDir string, dir *os.File, offline bool) (fis []fs.FileInfo, err error) {
	if fis, err = dir.Readdir(-1); err != nil {
		return
	}
	n := 0
	for _, fi := range fis {
		name := fi.Name()
		if strings.HasPrefix(name, SysFilePrefix) { // skip fscache system files
			continue
		}
		if isRemote(fi) {
			if offline {
				continue
			}
			localFile := filepath.Join(localDir, name)
			fi = readStubFile(localFile, fi)
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
	if fi.IsDir() {
		if checkDirCached(localFile) == nil { // no dir cache
			fi = &fileInfoRemote{fi}
		}
	} else if isRemote(fi) {
		fi = readStubFile(localFile, fi)
	}
	return
}

func (p *remote) SyncLstat(local string, name string) (fi fs.FileInfo, err error) {
	return nil, os.ErrNotExist
}

func (p *remote) SyncOpen(local string, name string) (f http.File, err error) {
	f, err = p.bucket.Open(name)
	if err != nil {
		log.Printf(`[ERROR] bucket.Open("%s"): %v\n`, name, err)
		return
	}
	if debugNet {
		log.Println("[INFO] ==> bucket.Open", name)
	}
	if f.(interface{ IsDir() bool }).IsDir() {
		fis, e := readdir(f)
		if e != nil {
			log.Printf(`[ERROR] Readdir("%s"): %v\n`, name, e)
			return nil, e
		}
		if debugNet {
			log.Println("[INFO] ==> Readdir", name, "-", len(fis), "items")
		}
		go func() {
			nError := 0
			base := filepath.Join(local, name)
			for _, fi := range fis {
				itemFile := base + "/" + fi.Name()
				if WriteStubFile(itemFile, fi) != nil {
					nError++
				}
			}
			if nError == 0 {
				TouchDirCached(base)
			} else {
				log.Printf("[WARN] writeStubFile fail (%d errors)", nError)
			}
		}()
		return cached.Dir(f, fis), nil
	}
	if p.cacheFile {
		localFile := filepath.Join(local, name)
		f = &objFile{f, localFile, p.notify}
	}
	return
}

func (p *remote) Init(local string, offline bool) {
}

// -----------------------------------------------------------------------------------------

type NotifyFile interface {
	NotifyFile(ctx context.Context, name string, fi fs.FileInfo)
}

func NewRemote(fsRemote http.FileSystem, notifyOrNil NotifyFile, cacheFile bool) (ret cached.Remote, err error) {
	return &remote{fsRemote, notifyOrNil, cacheFile}, nil
}

// NewCached creates a cached http.FileSystem to speed up listing directories and accessing file
// contents (optional, only when `cacheFile` is true). If `offline` is true, the cached http.FileSystem
// doesn't access `fsRemote http.FileSystem`.
func NewCached(local string, fsRemote http.FileSystem, notifyOrNil NotifyFile, cacheFile bool, offline ...bool) (fs http.FileSystem, err error) {
	r, err := NewRemote(fsRemote, notifyOrNil, cacheFile)
	if err != nil {
		return
	}
	return cached.New(local, r, offline...), nil
}

// -----------------------------------------------------------------------------------------
