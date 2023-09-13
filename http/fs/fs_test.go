package fs

import (
	"os"
	"testing"
)

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
