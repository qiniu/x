/*
 Copyright 2023 Qiniu Limited (qiniu.com)

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

package filter

import (
	"io/fs"
	"os"
	"testing"

	"github.com/qiniu/x/http/fs/fstest"
)

// -----------------------------------------------------------------------------------------

type caseMatched struct {
	patterns []string
	name     string
	dir      string
	fname    string
	isDir    bool
	matched  bool
}

func TestMatchted(t *testing.T) {
	cases := []caseMatched{
		{
			patterns: []string{"a.txt"},
			name:     "/foo/bar/a.txt",
			matched:  true,
		},
		{
			patterns: []string{"*.txt"},
			name:     "/foo/bar/a.txt",
			matched:  true,
		},
		{
			patterns: []string{"*~"},
			name:     "/foo/bar/a.txt~",
			matched:  true,
		},
		{
			patterns: []string{"a.*"},
			name:     "/foo/bar/a.txt",
			matched:  true,
		},
		{
			patterns: []string{"/*.txt"},
			name:     "/foo/bar/a.txt",
			matched:  false,
		},
		{
			patterns: []string{".*"},
			name:     "/foo/bar/.git",
			matched:  true,
		},
		{
			patterns: []string{"/foo"},
			name:     "/foo/bar/a.txt",
			matched:  false,
		},
		{
			patterns: []string{"/foo*"},
			name:     "/foo/bar/a.txt",
			matched:  true,
		},
		{
			patterns: []string{"/foo/"},
			name:     "/foo/bar/a.txt",
			matched:  true,
		},
		{
			patterns: []string{"/foo/bar/a.txt"},
			dir:      "/foo/bar",
			fname:    "a.txt",
			matched:  true,
		},
		{
			patterns: []string{"*."},
			dir:      "/foo/bar",
			fname:    "a.txt",
			matched:  false,
		},
		{
			patterns: []string{"*."},
			dir:      "/foo/bar",
			fname:    "abc",
			matched:  true,
		},
		{
			patterns: []string{"*."},
			dir:      "/foo/bar",
			fname:    "abc",
			isDir:    true,
			matched:  false,
		},
	}
	for _, c := range cases {
		if ret := Matched(c.patterns, c.name, c.dir, c.fname, c.isDir); ret != c.matched {
			t.Error("TestMatchted:", c.patterns, c.name, c.dir, c.fname, "expected:", c.matched, "ret:", ret)
		}
	}
}

// -----------------------------------------------------------------------------------------

func TestSelectRoot(t *testing.T) {
	if !Selected([]string{"a.txt"}, "/", true) {
		t.Fatal("Selected `a.txt`, `/`: fail")
	}
	if !Selected([]string{"/a/"}, "/", true) {
		t.Fatal("Selected `/a/`, `/`: fail")
	}
	if !Selected([]string{"/a.txt"}, "/", true) {
		t.Fatal("Selected `/a.txt`, `/`: fail")
	}
}

// -----------------------------------------------------------------------------------------

type openTestSel struct {
	name string
	n    int // n=1(one entry), n=0(no entry), n=-1(not exists)
}

type caseSelect struct {
	name     string
	patterns []string
	opens    []openTestSel
}

func TestSelect(t *testing.T) {
	cases := []caseSelect{
		{
			patterns: []string{"a.txt"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/", 1},
				{"/foo", 1},
				{"/foo/bar/a.txt", 1},
			},
		},
		{
			patterns: []string{"/a.txt"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/", 0},
				{"/foo", -1},
				{"/foo/bar/a.txt", -1},
			},
		},
		{
			patterns: []string{"/foo/"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/", 1},
				{"/foo", 1},
				{"/foo/bar", 1},
				{"/foo/bar/a.txt", 1},
			},
		},
		{
			patterns: []string{"/foo/bar/"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/", 1},
				{"/foo", 1},
				{"/foo/bar", 1},
				{"/foo/bar/a.txt", 1},
			},
		},
		{
			patterns: []string{"/foo*"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/", 1},
				{"/foo", 1},
				{"/foo/bar", 1},
				{"/foo/bar/a.txt", 1},
			},
		},
		{
			patterns: []string{"/foo/bar*"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/", 1},
				{"/foo", 1},
				{"/foo/bar", 1},
				{"/foo/bar/a.txt", 1},
			},
		},
		{
			patterns: []string{"/foo/bar/a.txt"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/foo", 1},
				{"/foo/bar", 1},
				{"/foo/bar/a.txt", 1},
			},
		},
		{
			patterns: []string{"/foo/bar"},
			name:     "/foo/bar/a.txt",
			opens: []openTestSel{
				{"/foo", 1},
				{"/foo/bar", 0},
				{"/foo/bar/a.txt", -1},
			},
		},
	}
	for _, c := range cases {
		fsys := Select(fstest.SingleFile(c.name, ""), c.patterns...)
		for _, o := range c.opens {
			f, err := fsys.Open(o.name)
			if o.n >= 0 {
				if err != nil {
					t.Error("TestSelect:", c.patterns, c.name, "- Open", o.name, "expected ret: nil, ret:", err)
				}
				if fi, err := f.Stat(); err == nil && fi.IsDir() {
					fis, e := f.Readdir(-1)
					if e != nil || len(fis) != o.n {
						t.Fatal("TestSelect:", c.patterns, c.name, o.name, len(fis), e)
					}
					f, _ = fsys.Open(o.name)
					items, e := f.(interface {
						ReadDir(count int) ([]fs.DirEntry, error)
					}).ReadDir(-1)
					if e != nil || len(items) != o.n {
						t.Error("TestSelect:", c.patterns, c.name, o.name, len(items), e)
					}
				}
			} else if !os.IsNotExist(err) {
				t.Error("TestSelect:", c.patterns, c.name, o.name, "expected: `file does not exist`, err:", err)
			}
		}
	}
}

// -----------------------------------------------------------------------------------------
