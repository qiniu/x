package gzip

import (
	"compress/gzip"
	"net/http"

	xfs "github.com/qiniu/x/http/fs"
)

func Open(fs http.FileSystem, name string) (file http.File, err error) {
	file, err = fs.Open(name)
	if err != nil {
		return
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	return xfs.SequenceFile(name, gr), nil
}
