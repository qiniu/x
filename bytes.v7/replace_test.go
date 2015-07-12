package bytes

import (
	"strings"
	"testing"
)

type replaceCase struct {
	s    string
	src  string
	dest string
	n    int
}

func stringReplace(b string, src, dest string, n int) string {

	return string(Replace([]byte(b), []byte(src), []byte(dest), n))
}

func TestReplace(t *testing.T) {

	cases := []replaceCase{
		{"hello, world!", "world", "xsw", -1},
		{"hello, world world world", "world", "xsw", 1},
		{"hello, world world world", "world", "xsw", -1},
		{"hello, xsw!", "xsw", "world", -1},
		{"hello, xsw xsw xsw", "xsw", "world", 1},
		{"hello, xsw xsw xsw", "xsw", "world", -1},
	}

	for _, c := range cases {
		ret := stringReplace(c.s, c.src, c.dest, c.n)
		expected := strings.Replace(c.s, c.src, c.dest, c.n)
		if ret != expected {
			t.Fatal("Replace failed:", c, "ret:", ret, "expected:", expected)
		}
	}
}

