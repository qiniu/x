//go:build unix
// +build unix

/*
Copyright 2019 Qiniu Limited (qiniu.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestBasicf(t *testing.T) {
	doSth := func(lvl int) {
		old := Std.Level
		defer func() {
			Std.Level = old
		}()
		Std.Level = lvl
		fns := []func(format string, v ...interface{}){
			Debugf, Infof, Warnf, Errorf,
		}
		for _, fn := range fns {
			fn("log %d\n", 100)
		}
	}
	doSth(Ldebug)
	doSth(Lfatal)
}

func TestBasic(t *testing.T) {
	doSth := func(lvl int) {
		old := Std.Level
		defer func() {
			Std.Level = old
		}()
		Std.Level = lvl
		fns := []func(v ...interface{}){
			Debug, Info, Warn, Error,
		}
		for _, fn := range fns {
			fn("log", 100)
		}
	}
	doSth(Ldebug)
	doSth(Lfatal)
}

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
	suffix := "x/log/logext_test.go:113: hello\n"
	if !strings.HasSuffix(result, suffix) {
		t.Errorf("Llongfile mode, logx.Info() = %v, which expected to include suffix %v ", result, suffix)
	}
}

func TestShortFile(t *testing.T) {
	out := bytes.Buffer{}
	logx := New(&out, Std.Prefix(), Lshortfile|Llevel|LstdFlags)
	logx.Info("hello")
	result := out.String()
	suffix := "logext_test.go:124: hello\n"
	if !strings.HasSuffix(result, suffix) {
		t.Errorf("Lshortfile mode, logx.Info() = %v, which expected to include suffix %v ", result, suffix)
	}
}
