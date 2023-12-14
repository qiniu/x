//go:build windows
// +build windows

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
		{"C:/log.v1/logext_test.go", 1, "logext_test.go"},
		{"C:/log.v1/logext_test.go", 2, "log.v1/logext_test.go"},
		{"github.com/log.v1/logext_test.go", 2, "github.com/log.v1/logext_test.go"},
	}
	for _, tc := range tcs {
		if actual := formatFile(tc.file, tc.lastN); actual != tc.expected {
			t.Errorf("formatFile(%v, %v) = %v, expected %v", tc.file, tc.lastN, actual, tc.expected)
		}
	}
}
