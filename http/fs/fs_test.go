package fs

import (
	"io"
	"os"
	"testing"
)

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
