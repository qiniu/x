//go:build unix
// +build unix

package log

import (
	"bytes"
	"regexp"
	"strings"
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

func TestLog(t *testing.T) {
	out := bytes.Buffer{}
	logx := New(&out, Std.Prefix(), Std.Flags())
	logx.Info("hello")
	result := out.String()

	// check time prefix
	actual := result[:26]
	if !regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}.\d{6}$`).MatchString(actual) {
		t.Errorf("time prefix of logx.Info() = %v, expected %v ", actual, "yyyy/mm/dd hh:mm:ss.xxxxxx")
	}

	// check level prefix
	actual = result[27:33]
	if actual != "[INFO]" {
		t.Errorf("level prefix of logx.Info() = %v, expected %v ", actual, "[INFO]")
	}

	// check file prefix
	actual = result[34:53]
	if actual != "log/logext_test.go:" {
		t.Errorf("file prefix of logx.Info() = %v, expected %v ", actual, " logext_test.go: ")
	}
}

func TestLongFile(t *testing.T) {
	out := bytes.Buffer{}
	logx := New(&out, Std.Prefix(), Llongfile|Llevel|LstdFlags)
	logx.Info("hello")
	result := out.String()
	suffix := "x/log/logext_test.go:61: hello\n"
	if !strings.HasSuffix(result, suffix) {
		t.Errorf("Llongfile mode, logx.Info() = %v, which expected to include suffix %v ", result, suffix)
	}
}

func TestShortFile(t *testing.T) {
	out := bytes.Buffer{}
	logx := New(&out, Std.Prefix(), Lshortfile|Llevel|LstdFlags)
	logx.Info("hello")
	result := out.String()
	suffix := "logext_test.go:72: hello\n"
	if !strings.HasSuffix(result, suffix) {
		t.Errorf("Lshortfile mode, logx.Info() = %v, which expected to include suffix %v ", result, suffix)
	}
}
