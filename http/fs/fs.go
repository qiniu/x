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

type unionFS struct {
	fs []http.FileSystem
}

func (p *unionFS) Open(name string) (f http.File, err error) {
	for _, fs := range p.fs {
		f, err = fs.Open(name)
		if !os.IsNotExist(err) {
			return
		}
	}
	return nil, fs.ErrNotExist
}

// Union merge a list of http.FileSystem into a union http.FileSystem object.
func Union(fs ...http.FileSystem) http.FileSystem {
	return &unionFS{fs}
}

// -----------------------------------------------------------------------------------------

type fsPlugins struct {
	fs   http.FileSystem
	exts map[string]Plugin
}

func (p *fsPlugins) Open(name string) (http.File, error) {
	ext := path.Ext(name)
	if fn, ok := p.exts[ext]; ok {
		return fn(p.fs, name)
	}
	return p.fs.Open(name)
}

type Plugin = func(fs http.FileSystem, name string) (file http.File, err error)

// Plugins implements a filesystem with plugins by specified (ext string, plugin Plugin) pairs.
func Plugins(fs http.FileSystem, plugins ...interface{}) http.FileSystem {
	n := len(plugins)
	exts := make(map[string]Plugin, n/2)
	for i := 0; i < n; i += 2 {
		ext := plugins[i].(string)
		fn := plugins[i+1].(Plugin)
		exts[ext] = fn
	}
	return &fsPlugins{fs, exts}
}

// -----------------------------------------------------------------------------------------

// FileInfo describes a single file in an file system.
// It implements fs.FileInfo and fs.DirEntry.
type FileInfo struct {
	name  string
	size  int64
	Mtime time.Time
}

// NewFileInfo creates a FileInfo that describes a single file in an file system.
// It implements fs.FileInfo and fs.DirEntry.
func NewFileInfo(name string, size int64) *FileInfo {
	return &FileInfo{name: name, size: size}
}

// for fs.FileInfo, fs.DirEntry
func (p *FileInfo) Name() string {
	return p.name
}

// for fs.FileInfo
func (p *FileInfo) Size() int64 {
	return p.size
}

// for fs.FileInfo
func (p *FileInfo) Mode() fs.FileMode {
	return fs.ModeIrregular
}

// fs.DirEntry
func (p *FileInfo) Type() fs.FileMode {
	return fs.ModeIrregular
}

// for fs.FileInfo
func (p *FileInfo) ModTime() time.Time {
	return p.Mtime
}

// for fs.FileInfo, fs.DirEntry
func (p *FileInfo) IsDir() bool {
	return false
}

// fs.DirEntry
func (p *FileInfo) Info() (fs.FileInfo, error) {
	return p, nil
}

// for fs.FileInfo
func (p *FileInfo) Sys() interface{} {
	return nil
}

// -----------------------------------------------------------------------------------------

// DirInfo describes a single directory in an file system.
// It implements fs.FileInfo and io.Closer and Stat().
type DirInfo struct {
	name string
}

// NewDirInfo creates a DirInfo that describes a single directory in an file system.
// It implements fs.FileInfo and io.Closer and Stat().
func NewDirInfo(name string) *DirInfo {
	return &DirInfo{name}
}

// for fs.FileInfo, fs.DirEntry
func (p *DirInfo) Name() string {
	return p.name
}

// for fs.FileInfo
func (p *DirInfo) Size() int64 {
	return 0
}

// for fs.FileInfo
func (p *DirInfo) Mode() fs.FileMode {
	return fs.ModeIrregular | fs.ModeDir
}

// for fs.FileInfo
func (p *DirInfo) ModTime() time.Time {
	return time.Now()
}

// for fs.FileInfo, fs.DirEntry
func (p *DirInfo) IsDir() bool {
	return true
}

// for fs.FileInfo
func (p *DirInfo) Sys() interface{} {
	return nil
}

// for fs.File, http.File
func (p *DirInfo) Stat() (fs.FileInfo, error) {
	return p, nil
}

// io.Closer
func (p *DirInfo) Close() error {
	return nil
}

// -----------------------------------------------------------------------------------------

type rootDir struct {
}

func (p rootDir) Name() string {
	return "/"
}

func (p rootDir) Size() int64 {
	return 0
}

func (p rootDir) Mode() fs.FileMode {
	return fs.ModeDir
}

func (p rootDir) ModTime() time.Time {
	return time.Now()
}

func (p rootDir) IsDir() bool {
	return true
}

func (p rootDir) Sys() interface{} {
	return nil
}

func (p rootDir) Close() error {
	return nil
}

func (p rootDir) Write(b []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (p rootDir) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

func (p rootDir) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, io.EOF
}

func (p rootDir) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, io.EOF
}

func (p rootDir) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart && offset == 0 {
		return 0, nil
	}
	return 0, io.EOF
}

func (p rootDir) Stat() (fs.FileInfo, error) {
	return rootDir{}, nil
}

func (p rootDir) Open(name string) (f http.File, err error) {
	if name == "/" {
		return rootDir{}, nil
	}
	return nil, fs.ErrNotExist
}

// Root implements a http.FileSystem that only have a root directory.
func Root() http.FileSystem {
	return rootDir{}
}

// -----------------------------------------------------------------------------------------

type parentFS struct {
	parentDir string
	fs        http.FileSystem
}

func Parent(parentDir string, fs http.FileSystem) http.FileSystem {
	parentDir = strings.TrimSuffix(parentDir, "/")
	if !strings.HasPrefix(parentDir, "/") {
		parentDir = "/" + parentDir
	}
	return &parentFS{parentDir, fs}
}

func (p *parentFS) Open(name string) (f http.File, err error) {
	if !strings.HasPrefix(name, p.parentDir) {
		return nil, os.ErrNotExist
	}
	path := name[len(p.parentDir):]
	return p.fs.Open(path)
}

// -----------------------------------------------------------------------------------------

type subFS struct {
	subDir string
	fs     http.FileSystem
}

func Sub(fs http.FileSystem, subDir string) http.FileSystem {
	subDir = strings.TrimSuffix(subDir, "/")
	if !strings.HasPrefix(subDir, "/") {
		subDir = "/" + subDir
	}
	return &subFS{subDir, fs}
}

func (p *subFS) Open(name string) (f http.File, err error) {
	return p.fs.Open(p.subDir + name)
}

// -----------------------------------------------------------------------------------------
