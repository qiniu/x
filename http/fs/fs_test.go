package fs

import "testing"

func TestParentFS(t *testing.T) {
	fs := FilesWithContent("a.txt", "a")
	if _, err := fs.Open("/a.txt"); err != nil {
		t.Fatal(err)
	}

	p := Parent("files", fs)
	if _, err := p.Open("/files/a.txt"); err != nil {
		t.Fatal(err)
	}
}

func TestSubFS(t *testing.T) {
	fs := FilesWithContent("files/a.txt", "a")
	if _, err := fs.Open("/files/a.txt"); err != nil {
		t.Fatal(err)
	}

	s := Sub("files", fs)
	if _, err := s.Open("/a.txt"); err != nil {
		t.Fatal(err)
	}
}
