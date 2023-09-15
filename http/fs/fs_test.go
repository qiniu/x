package fs

import (
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
	testFile(t, Http("http://a.com", context.TODO()).With(mockClient, nil), &expectFile{name: "/", readdirErr: fs.ErrInvalid, readN: 1})
	testFile(t, Http("http://b.com").With(mockClient, nil), &expectFile{name: "/", openErr: &fs.PathError{Op: "http.Get", Path: "http://b.com/", Err: fs.ErrNotExist}})
	testFile(t, Http("http://c.com").With(mockClient, nil), &expectFile{name: "/", openErr: &fs.PathError{Op: "http.Get", Path: "http://c.com/", Err: err500}})
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
	testFI(t, HttpFile("/foo/a.txt", &http.Response{ContentLength: -1}).(*httpFile), "a.txt", -1, fs.ModeIrregular, false, false, nil)
	testFI(t, SequenceFile("/foo/a.txt", nil).(*stream), "a.txt", -1, fs.ModeIrregular, false, false, nil)
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
