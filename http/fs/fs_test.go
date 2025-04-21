/*
 Copyright 2024 Qiniu Limited (qiniu.com)

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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/x/mockhttp"
	"github.com/qiniu/x/ts"
)

// -----------------------------------------------------------------------------------------

type iStream interface {
	http.File
	TryReader() *bytes.Reader
	FullName() string
	Size() int64
	ModTime() time.Time
}

type expectStm struct {
	br       *bytes.Reader
	fullName string
	readN    int
	readErr  error
	size     int64
	modt     time.Time
	unseek   io.ReadCloser
}

type sizeModt struct {
	errCloser
}

func (sizeModt) Size() int64 {
	return 123
}

func (sizeModt) ModTime() time.Time {
	return time.Time{}
}

func testIS(t *testing.T, stm iStream, exp *expectStm) {
	ts := ts.New(t)
	ts.Case("TryReader", stm.TryReader()).Equal(exp.br)
	ts.Case("FullName", stm.FullName()).Equal(exp.fullName)
	ts.New("stm.Read").Init(stm.Read(make([]byte, 1))).Equal(exp.readN, exp.readErr)
	ts.Case("stm.Size", stm.Size()).Equal(exp.size)
	ts.Case("stm.ModTime", stm.ModTime()).Equal(exp.modt)
	ts.Case("Unseekable", Unseekable(stm) != exp.unseek).Equal(true)
}

func TestStream(t *testing.T) {
	br := bytes.NewReader([]byte("a"))
	file := &sizeModt{}
	stm := &stream{br: br, file: file}
	testIS(t, stm, &expectStm{br: br, readN: 1, size: 123})
	br.Seek(0, io.SeekStart)
	ti := time.Now().Format(http.TimeFormat)
	stm2 := &httpFile{stream: stream{br: br, file: file}, resp: &http.Response{ContentLength: 123, Body: file, Header: http.Header{
		"Last-Modified": []string{ti},
	}}}
	modt, err := http.ParseTime(ti)
	if err != nil {
		t.Fatal("http.ParseTime:", err)
	}
	testIS(t, stm2, &expectStm{br: br, readN: 1, size: 123, modt: modt})
	stm.br = nil
	if f := Unseekable(stm); f != file {
		t.Fatal("Unseekable:", f)
	}
	stm2.br = nil
	if f := Unseekable(stm2); f != file {
		t.Fatal("Unseekable:", f)
	}
}

func TestCopyFile(t *testing.T) {
	f := HttpFile("/foo/a.txt", &http.Response{ContentLength: -1, Body: io.NopCloser(strings.NewReader("abc"))})
	if err := CopyFile(io.Discard, f); err != nil {
		t.Fatal("CopyFile:", err)
	}
	f.Seek(0, io.SeekStart)
	if err := CopyFile(io.Discard, f); err != nil {
		t.Fatal("CopyFile:", err)
	}
}

type iDirReader interface {
	http.File
	ReadDir(n int) (items []fs.DirEntry, err error)
}

func testReadDir(t *testing.T, d iDirReader, n, nexp int) {
	des, err := d.ReadDir(n)
	if err != nil {
		if n < 2 || err != io.EOF {
			t.Fatal("Dir.Readdir:", err)
		}
	}
	if len(des) != nexp {
		t.Fatal("Dir.Readdir len(des):", len(des))
	}
	if des[0].Name() != "a.txt" {
		t.Fatal("Dir.Readdir fis[0].Name():", des[0].Name())
	}
	des[0].Type()
	des[0].Info()
}

func TestDir(t *testing.T) {
	{
		fis := []fs.FileInfo{NewFileInfo("a.txt", 123), NewFileInfo("b.txt", 456)}
		d := Dir(NewDirInfo(""), fis).(iDirReader)
		testReadDir(t, d, -1, 2)
		d.Seek(0, io.SeekStart)
		testReadDir(t, d, 5, 2)
	}
	{
		fi := NewFileInfo("a.txt", 123)
		fis := []fs.FileInfo{fi, NewFileInfo("b.txt", 456)}
		d := Dir(NewDirInfo(""), fis).(iDirReader)
		testReadDir(t, d, 1, 1)
		d.Seek(0, io.SeekEnd)
		d.Read(nil)
		d.Stat()
		d.Close()
		fi.Info()
		fi.Type()
	}
}

// -----------------------------------------------------------------------------------------

type errCloser struct {
	io.Reader
}

func (p *errCloser) Close() error {
	return fs.ErrPermission
}

type statCloser interface {
	io.Closer
	Stat() (fs.FileInfo, error)
}

type dirReader interface {
	ReadDir(n int) ([]fs.DirEntry, error)
}

type expectFile struct {
	name       string
	close      error
	openErr    error
	statErr    error
	readdirFIs []fs.FileInfo
	readdirDEs []fs.DirEntry
	readdirErr error
	readErr    error
	seekOff    int64
	seekNewOff int64
	seekErr    error
	seekWhence int
	readN      int
}

func testSC(ts *ts.Testing, f statCloser, exp *expectFile) {
	ts.Case("f.Close", f.Close()).Equal(exp.close)
	ts.New("f.Stat").Init(f.Stat()).Next().Equal(exp.statErr)
}

func testHF(ts *ts.Testing, f http.File, exp *expectFile) {
	testSC(ts, f, exp)
	ts.New("f.Read").Init(f.Read(make([]byte, 1))).Equal(exp.readN, exp.readErr)
	ts.New("f.Readdir").Init(f.Readdir(-1)).Equal(exp.readdirFIs, exp.readdirErr)
	ts.New("f.ReadDir").Init(f.(dirReader).ReadDir(-1)).Equal(exp.readdirDEs, exp.readdirErr)
	ts.New("f.Seek").Init(f.Seek(exp.seekOff, exp.seekWhence)).Equal(exp.seekNewOff, exp.seekErr)
}

func testFile(t *testing.T, v interface{}, exp *expectFile) {
	ts := ts.New(t)
	switch f := v.(type) {
	case http.FileSystem:
		file, err := f.Open(exp.name)
		ts.New("f.Open").Init(err).Equal(exp.openErr)
		if err == nil {
			testHF(ts, file, exp)
		}
	case http.File:
		testHF(ts, f, exp)
	case statCloser:
		testSC(ts, f, exp)
	}
}

func TestHttpFile(t *testing.T) {
	testFile(t, FilesWithContent("foo/a.txt", "a"), &expectFile{name: "/foo/a.txt", readdirErr: fs.ErrInvalid, readN: 1})
	testFile(t, Root(), &expectFile{name: "/", readdirErr: io.EOF, readErr: io.EOF})
	mock.ListenAndServe("a.com", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "a")
	}))
	mock.ListenAndServe("b.com", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	mock.ListenAndServe("c.com", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	err500 := fmt.Errorf("http.Get %s error: status %d (%s)", "http://c.com/", 500, "")
	aCom := Http("http://a.com", context.TODO()).With(mockClient, nil)
	bCom := Http("http://b.com").With(mockClient, nil)
	cCom := Http("http://c.com").With(mockClient, nil)
	testFile(t, aCom, &expectFile{name: "/", readdirErr: fs.ErrInvalid, readN: 1})
	testFile(t, bCom, &expectFile{name: "/", openErr: &fs.PathError{Op: "http.Get", Path: "http://b.com/", Err: fs.ErrNotExist}})
	testFile(t, cCom, &expectFile{name: "/", openErr: &fs.PathError{Op: "http.Get", Path: "http://c.com/", Err: err500}})
	track := WithTracker(Root(), "http://a.com", ".txt").(*fsWithTracker)
	track.httpfs = aCom
	testFile(t, track, &expectFile{name: "/foo.txt", readdirErr: fs.ErrInvalid, readN: 1})
	testFile(t, WithTracker(bCom, aCom, ".txt"), &expectFile{name: "/bar.jpg", openErr: &fs.PathError{Op: "http.Get", Path: "http://b.com/bar.jpg", Err: fs.ErrNotExist}})
	testFile(t, SequenceFile("/foo/bar.txt", &errCloser{strings.NewReader("a")}), &expectFile{close: fs.ErrPermission, readdirErr: fs.ErrInvalid, readN: 1})
}

var (
	mockClient = mockhttp.DefaultClient
	mock       = mockhttp.DefaultTransport
)

// -----------------------------------------------------------------------------------------

func testFI(t *testing.T, fi fs.FileInfo, name string, size int64, mode fs.FileMode, zeroTime, isDir bool, sys interface{}) {
	ts := ts.New(t)
	ts.Case("fi.Name", fi.Name()).Equal(name)
	ts.Case("fi.Size", fi.Size()).Equal(size)
	ts.Case("fi.Mode", fi.Mode()).Equal(mode)
	if zeroTime {
		ts.Case("fi.ModTime", fi.ModTime()).Equal(time.Time{})
	} else if mt := fi.ModTime(); mt.IsZero() {
		t.Fatal("fi.ModTime:", mt)
	}
	ts.Case("fi.IsDir", fi.IsDir()).Equal(isDir)
	ts.Case("fi.Sys", fi.Sys()).Equal(sys)
}

func TestFileInfo(t *testing.T) {
	testFI(t, NewDirInfo("foo"), "foo", 0, fs.ModeIrregular|fs.ModeDir, false, true, nil)
	testFI(t, NewFileInfo("a.txt", 123), "a.txt", 123, fs.ModeIrregular, true, false, nil)
	testFI(t, rootDir{}, "/", 0, fs.ModeDir, false, true, nil)
	testFI(t, &dataFile{name: "/foo/a.txt", ContentReader: strings.NewReader("a")}, "a.txt", 1, 0, false, false, nil)
	testFI(t, HttpFile("/foo/a.txt", &http.Response{ContentLength: -1, Body: io.NopCloser(strings.NewReader("abc"))}).(*httpFile), "a.txt", 3, fs.ModeIrregular, false, false, nil)
	testFI(t, SequenceFile("/foo/a.txt", io.NopCloser(strings.NewReader("abc"))).(*stream), "a.txt", 3, fs.ModeIrregular, false, false, nil)
}

// -----------------------------------------------------------------------------------------

func TestUnionFS(t *testing.T) {
	fs := Union(Root(), Files("fs/file.go", "/not-exist/noop.nil"), FilesWithContent("fs/file.go", "a"))
	f, err := fs.Open("/fs/file.go")
	if err != nil {
		t.Fatal("Union.Open:", err)
	}
	if b, err := io.ReadAll(f); err != nil || string(b) != "a" {
		t.Fatal("UnionFS ReadAll:", string(b), err)
	}
	_, err = fs.Open("/not-exist/foo")
	if !os.IsNotExist(err) {
		t.Fatal("fs.Open:", err)
	}
}

func TestParentFS(t *testing.T) {
	fs := FilesWithContent("a.txt", "a")
	if _, err := fs.Open("/a.txt"); err != nil {
		t.Fatal(err)
	}

	p := Parent("files", fs)
	if _, err := p.Open("/files/a.txt"); err != nil {
		t.Fatal(err)
	}
	_, err := p.Open("/a.txt")
	if err == nil {
		t.Fatal("expect error")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("expect os.IsNotExist(err) but got %v", err)
	}
}

func TestSubFS(t *testing.T) {
	fs := FilesWithContent("files/a.txt", "a")
	if _, err := fs.Open("/files/a.txt"); err != nil {
		t.Fatal(err)
	}

	s := Sub(fs, "files")
	if _, err := s.Open("/a.txt"); err != nil {
		t.Fatal(err)
	}
}

// -----------------------------------------------------------------------------------------
