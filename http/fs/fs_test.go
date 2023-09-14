package fs

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/x/ts"
)

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
