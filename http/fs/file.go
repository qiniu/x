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

package fs

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------------------

type ContentReader interface {
	io.Reader
	io.Seeker
	Size() int64
}

type dataFile struct {
	ContentReader
	name string
}

func (p *dataFile) Close() error {
	if r, ok := p.ContentReader.(io.Closer); ok {
		return r.Close()
	}
	return nil
}

func (p *dataFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, fs.ErrInvalid
}

func (p *dataFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (p *dataFile) IsDir() bool {
	return false
}

func (p *dataFile) Mode() fs.FileMode {
	return 0
}

func (p *dataFile) Stat() (fs.FileInfo, error) {
	return p, nil
}

func (p *dataFile) Name() string {
	return path.Base(p.name)
}

func (p *dataFile) FullName() string {
	return p.name
}

func (p *dataFile) Size() int64 {
	return p.ContentReader.Size()
}

func (p *dataFile) ModTime() time.Time {
	if r, ok := p.ContentReader.(interface{ ModTime() time.Time }); ok {
		return r.ModTime()
	}
	return time.Now()
}

func (p *dataFile) Sys() interface{} {
	return nil
}

// File implements a http.File by a ContentReader which may implement
// optional interface{ ModTime() time.Time } and io.Closer.
func File(name string, r ContentReader) http.File {
	return &dataFile{r, name}
}

// -----------------------------------------------------------------------------------------

type filesDataFS struct {
	files []string
}

func (p *filesDataFS) Open(name string) (f http.File, err error) {
	files := p.files
	name = name[1:]
	for i := 0; i < len(files); i += 2 {
		if files[i] == name {
			return File(name, strings.NewReader(files[i+1])), nil
		}
	}
	return nil, fs.ErrNotExist
}

// FilesWithContent implements a http.FileSystem by a list of file name and content.
func FilesWithContent(files ...string) http.FileSystem {
	return &filesDataFS{files}
}

// -----------------------------------------------------------------------------------------

type filesFS struct {
	files []string
}

func (p *filesFS) Open(name string) (f http.File, err error) {
	files := p.files
	name = name[1:]
	for i := 0; i < len(files); i += 2 {
		if files[i] == name {
			f, err = os.Open(files[i+1])
			return
		}
	}
	return nil, fs.ErrNotExist
}

// Files implements a http.FileSystem by a list of file name and content file.
func Files(files ...string) http.FileSystem {
	return &filesFS{files}
}

// -----------------------------------------------------------------------------------------
