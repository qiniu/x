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

	xfs "github.com/qiniu/x/http/fs"
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
	SysFilePrefix = ".fscache."

	// readDir from local is ready if there is a dirListCacheFile
	// see MarkDirCached
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

// MarkDirCached marks the directory has dirList cache.
func MarkDirCached(dir string) error {
	cacheFile := filepath.Join(dir, dirListCacheFile)
	return os.WriteFile(cacheFile, nil, 0666)
}

// WriteStubFile writes a stub file to the local file system.
// If the file is a directory, it creates corresponding directory.
func WriteStubFile(localFile string, fi fs.FileInfo, udata uint64) error {
	if fi.IsDir() {
		// directory don't need to save FileInfo
		return os.Mkdir(localFile, 0755)
	}
	fi = &fileInfoRemote{fi}
	b := xdir.BytesFileInfo(fi, udata)
	dest := base64.URLEncoding.EncodeToString(b)
	// don't need to restore mtime for symlink (saved by BytesFileInfo)
	return os.Symlink(dest, localFile)
}

// Udata returns the user data from the FileInfo.
func Udata(fi fs.FileInfo) uint64 {
	if fi, ok := fi.(interface{ Udata() uint64 }); ok {
		return fi.Udata()
	}
	return 0
}

// MkStubFile creates a stub file with the specified name, size and mtime.
func MkStubFile(rootDir string, name string, size int64, mtime time.Time, udata uint64) (err error) {
	file := filepath.Join(rootDir, name)
	dir, fname := filepath.Split(file)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}
	fi := xfs.NewFileInfo(fname, size)
	fi.Mtime = mtime
	return WriteStubFile(file, fi, udata)
}

// ReaddirAll reads all entries in the directory and returns their cached FileInfo.
func ReaddirAll(localDir string, dir *os.File, offline bool) (fis []fs.FileInfo, err error) {
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

// Lstat returns the FileInfo for the specified file or directory.
func Lstat(localFile string) (fi fs.FileInfo, err error) {
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

/* TODO(xsw): cache file
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
*/

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
	return ReaddirAll(localDir, dir, offline)
}

func (p *remote) Lstat(localFile string) (fi fs.FileInfo, err error) {
	return Lstat(localFile)
}

func (p *remote) SyncLstat(local string, name string) (fi fs.FileInfo, err error) {
	return nil, os.ErrNotExist
}

func (p *remote) SyncOpen(local string, name string, fi fs.FileInfo) (f http.File, err error) {
	f, err = p.bucket.Open(name)
	if err != nil {
		log.Printf(`[ERROR] bucket.Open("%s"): %v\n`, name, err)
		return
	}
	if debugNet {
		log.Println("[INFO] ==> bucket.Open", name)
	}
	if fi.IsDir() {
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
				if WriteStubFile(itemFile, fi, 0) != nil {
					nError++
				}
			}
			if nError == 0 {
				MarkDirCached(base)
			} else {
				log.Printf("[WARN] writeStubFile fail (%d errors)", nError)
			}
		}()
		return xfs.Dir(f, fis), nil
	}
	/* TODO(xsw):
	if p.cacheFile {
		localFile := filepath.Join(local, name)
		f = &objFile{f, localFile, p.notify}
	}
	*/
	return
}

func (p *remote) Init(local string, offline bool) error {
	return nil
}

// -----------------------------------------------------------------------------------------

type NotifyFile interface {
	NotifyFile(ctx context.Context, name string, fi fs.FileInfo)
}

// NewRemote creates a new remote file system.
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
