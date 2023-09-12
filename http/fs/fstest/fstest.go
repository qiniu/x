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

package fstest

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"reflect"
	"strings"

	xfs "github.com/qiniu/x/http/fs"
	"github.com/qiniu/x/http/fs/cached"
)

// -----------------------------------------------------------------------------------------

// Dir creates a file http.File and its content is `data`.
func File(name, data string) http.File {
	return xfs.File(name, strings.NewReader(data))
}

// -----------------------------------------------------------------------------------------

type fsDir struct {
	http.File
	entries []http.File
}

func (p *fsDir) ReadDir(count int) ([]fs.DirEntry, error) {
	if f, ok := p.File.(interface {
		ReadDir(count int) ([]fs.DirEntry, error)
	}); ok {
		return f.ReadDir(count)
	}
	return nil, &fs.PathError{Op: "readdir", Err: errors.New("not implemented")}
}

// Dir creates a directory http.File and its entry can be File(name, data) or Dir(name, entries).
func Dir(name string, entries ...http.File) http.File {
	fi := xfs.NewDirInfo(name)
	fis := make([]fs.FileInfo, len(entries))
	for i, entry := range entries {
		item, e := entry.Stat()
		if e != nil {
			panic(e)
		}
		fis[i] = item
	}
	return &fsDir{cached.Dir(fi, fis), entries}
}

// -----------------------------------------------------------------------------------------

type fsMap struct {
	items map[string]http.File
}

func (p fsMap) Open(name string) (http.File, error) {
	if f, ok := p.items[name]; ok {
		_, e := f.Seek(0, io.SeekStart)
		if e != nil {
			log.Panicln("file doesn't support `Seek`:", reflect.TypeOf(f))
		}
		return f, nil
	}
	return nil, fs.ErrNotExist
}

func makeMap(ret map[string]http.File, dir string, entries []http.File) {
	for _, entry := range entries {
		item, e := entry.Stat()
		if e != nil {
			panic(e)
		}
		name := dir + item.Name()
		ret[name] = entry
		if f, ok := entry.(*fsDir); ok {
			makeMap(ret, name+"/", f.entries)
		}
	}
}

// FS creates a file system for testing. An entry can be File(name, data) or Dir(name, entries).
func FS(entries ...http.File) http.FileSystem {
	ret := make(map[string]http.File, 4)
	makeMap(ret, "/", entries)
	ret["/"] = Dir("/", entries...)
	return fsMap{ret}
}

// -----------------------------------------------------------------------------------------

// Single creates a file system that only contains a signle file (but it may not in the root directory).
func SingleFile(path, data string) http.FileSystem {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	n := len(parts)
	f := File(parts[n-1], data)
	for i := n - 2; i >= 0; i-- {
		f = Dir(parts[i], f)
	}
	return FS(f)
}

// -----------------------------------------------------------------------------------------
