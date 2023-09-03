package dircache

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	dircache "github.com/qiniu/x/http/fs/cached/dir"
)

const (
	CacheFileName string = ".cache"
)

func remoteStat(localFile string, fi fs.FileInfo) fs.FileInfo
func isRemote(fi fs.FileInfo, file string) bool

func ReaddirAll(localDir string, dir *os.File) (fis []fs.FileInfo, err error) {
	cacheFile := filepath.Join(localDir, CacheFileName)
	b, err := os.ReadFile(cacheFile)
	if err == nil {
		fis, err = dircache.ReadFileInfos(b)
		if err == nil {
			return
		}
		log.Printf("ReaddirAll: %s cache file broken and skipped - %v\n", localDir, err)
	}
	if fis, err = dir.Readdir(-1); err != nil {
		return
	}
	data := make([]byte, dircache.SizeFileInfos(fis))
	b = dircache.WriteCacheHdr(data, len(fis))
	for i, fi := range fis {
		name := fi.Name()
		if isRemote(fi, name) {
			localFile := filepath.Join(localDir, name)
			fi = remoteStat(localFile, fi)
			fis[i] = fi
		}
		b = dircache.WriteFileInfo(b, fi)
	}
	os.WriteFile(cacheFile, data, 0666)
	return
}
