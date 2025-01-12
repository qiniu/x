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

func TestBuild(t *testing.T) {
	if ret := NewBuilder(nil).Build(); ret != "" {
		t.Fatal("NewBuilder(0):", ret)
	}
	if ret := NewBuilderSize(16).Add("1").AddByte('2', '3').AddByte('!').Build(); ret != "123!" {
		t.Fatal("TestBuild:", ret)
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
