//go:build unix
// +build unix

package log

import (
	"testing"
)

func TestFormatFile(t *testing.T) {
	tcs := []struct {
		file     string
		lastN    int
		expected string
	}{
		{"github.com/qiniu/log.v1/logext_test.go", 1, "github.com/qiniu/log.v1/logext_test.go"},
		{"/qbox/github.com/qiniu/log.v1/logext_test.go", 2, "log.v1/logext_test.go"},
		{"/qbox/github.com/qiniu/log.v1/logext_test.go", 1, "logext_test.go"},
		{"logext_test.go", 2, "logext_test.go"},
		{"/logext_test.go", 2, "/logext_test.go"},
		{"/a/logext_test.go", 2, "a/logext_test.go"},
	}
	for _, tc := range tcs {
		if actual := formatFile(tc.file, tc.lastN); actual != tc.expected {
			t.Errorf("formatFile(%v, %v) = %v, expected %v", tc.file, tc.lastN, actual, tc.expected)
		}
	}
}
