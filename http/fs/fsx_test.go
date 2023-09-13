package fs_test

import (
	"net/http"
	"testing"

	"github.com/qiniu/x/http/fs"
	"github.com/qiniu/x/http/fs/filter"
)

func TestLocalCheck(t *testing.T) {
	fsys := http.Dir("/")
	if dir, ok := fs.LocalCheck(fsys); !ok || dir != "/" {
		t.Fatal("fs.LocalCheck(http.Dir):", dir, ok)
	}
	selFs := filter.Select(fsys, "*.txt")
	if dir, ok := fs.LocalCheck(selFs); !ok || dir != "/" {
		t.Fatal("fs.LocalCheck(filterFS):", dir, ok)
	}
	if _, ok := fs.LocalCheck(fs.Root()); ok {
		t.Fatal("fs.LocalCheck(root):", ok)
	}
}
