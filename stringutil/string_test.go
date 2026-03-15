/*
 Copyright 2024 Qiniu Limited (qiniu.com)

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

package stringutil

import (
	"reflect"
	"sort"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name      string
		classAttr string
		classVal  string
		want      bool
	}{
		// Basic matches
		{
			name:      "matches first class",
			classAttr: "param-required required",
			classVal:  "param-required",
			want:      true,
		},
		{
			name:      "matches last class",
			classAttr: "param-required required",
			classVal:  "required",
			want:      true,
		},
		{
			name:      "matches sole class",
			classAttr: "foo",
			classVal:  "foo",
			want:      true,
		},
		{
			name:      "matches middle class",
			classAttr: "foo bar baz",
			classVal:  "bar",
			want:      true,
		},

		// Partial / substring must not match
		{
			name:      "rejects prefix substring",
			classAttr: "param-required required",
			classVal:  "param",
			want:      false,
		},
		{
			name:      "rejects suffix substring",
			classAttr: "param-required required",
			classVal:  "required-extra",
			want:      false,
		},
		{
			name:      "rejects substring in the middle",
			classAttr: "foo bar baz",
			classVal:  "ba",
			want:      false,
		},

		// Edge cases
		{
			name:      "empty classAttr",
			classAttr: "",
			classVal:  "foo",
			want:      false,
		},
		{
			name:      "empty classVal",
			classAttr: "foo bar",
			classVal:  "",
			want:      false,
		},
		{
			name:      "both empty",
			classAttr: "",
			classVal:  "",
			want:      false,
		},
		{
			name:      "leading spaces",
			classAttr: "   foo bar",
			classVal:  "foo",
			want:      true,
		},
		{
			name:      "trailing spaces",
			classAttr: "foo bar   ",
			classVal:  "bar",
			want:      true,
		},
		{
			name:      "multiple spaces between classes",
			classAttr: "foo   bar   baz",
			classVal:  "bar",
			want:      true,
		},
		{
			name:      "class not present",
			classAttr: "foo bar baz",
			classVal:  "qux",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.classAttr, tt.classVal)
			if got != tt.want {
				t.Errorf("Contains(%q, %q) = %v, want %v",
					tt.classAttr, tt.classVal, got, tt.want)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	if ret := Capitalize(""); ret != "" {
		t.Fatal("Capitalize:", ret)
	}
	if ret := Capitalize("hello"); ret != "Hello" {
		t.Fatal("Capitalize:", ret)
	}
	if ret := Capitalize("Hello"); ret != "Hello" {
		t.Fatal("Capitalize:", ret)
	}
}

func TestConcat(t *testing.T) {
	if ret := Concat("1"); ret != "1" {
		t.Fatal("Concat(1):", ret)
	}
	if ret := Concat("1", "23", "!"); ret != "123!" {
		t.Fatal("Concat:", ret)
	}
}

func TestDiff(t *testing.T) {
	type testCase struct {
		new, old []string
		add, del []string
	}
	cases := []testCase{
		{[]string{"1", "3", "2", "4"}, []string{"2"}, []string{"1", "3", "4"}, nil},
		{[]string{"1", "3", "2", "4"}, []string{"5", "2"}, []string{"1", "3", "4"}, []string{"5"}},
		{[]string{"1", "3", "2", "4"}, []string{"0", "5", "2"}, []string{"1", "3", "4"}, []string{"0", "5"}},
	}
	for _, c := range cases {
		add, del := uDiff(c.new, c.old)
		if !reflect.DeepEqual(add, c.add) || !reflect.DeepEqual(del, c.del) {
			t.Fatal("diff:", c, "=>", add, del)
		}
	}
}

func uDiff(new, old []string) (add, del []string) {
	sort.Strings(new)
	sort.Strings(old)
	return Diff(new, old)
}
