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

package ctype

import (
	"testing"
)

type testCase struct {
	c    rune
	mask uint32
	is   bool
}

type stringTestCase struct {
	str       string
	maskFirst uint32
	maskNext  uint32
	is        bool
}

var isCases = []testCase{
	{'-', DOMAIN_CHAR, true},
	{'.', DOMAIN_CHAR, true},
	{'_', DOMAIN_CHAR, false},
	{'+', DOMAIN_CHAR, true},
	{'a', DOMAIN_CHAR, true},
	{'A', DOMAIN_CHAR, true},
	{'0', DOMAIN_CHAR, true},
	{':', DOMAIN_CHAR, false},
	{'1', ALPHA, false},
	{'a', ALPHA, true},
	{'A', ALPHA, true},
}

var strCases = []stringTestCase{
	{"", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, false},
	{"123", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, false},
	{"_", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, true},
	{"_123", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, true},
	{"x_123", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, true},
	{"x_", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, true},
	{"_x", CSYMBOL_FIRST_CHAR, CSYMBOL_NEXT_CHAR, true},

	{"", CSYMBOL_FIRST_CHAR, CSYMBOL_FIRST_CHAR, false},
	{"x_123", CSYMBOL_FIRST_CHAR, CSYMBOL_FIRST_CHAR, false},
	{"x_", CSYMBOL_FIRST_CHAR, CSYMBOL_FIRST_CHAR, true},
	{"_x", CSYMBOL_FIRST_CHAR, CSYMBOL_FIRST_CHAR, true},
	{"_", CSYMBOL_FIRST_CHAR, CSYMBOL_FIRST_CHAR, true},
}

func TestIs(t *testing.T) {
	if Is(0, 255) {
		t.Fatal("Is(0, 255)")
	}
	for _, a := range isCases {
		f := Is(a.mask, a.c)
		if f != a.is {
			t.Fatal("case:", a, "result:", f)
		}
	}
}

func TestIsTypeEx(t *testing.T) {
	for _, a := range strCases {
		f := IsTypeEx(a.maskFirst, a.maskNext, a.str)
		if f != a.is {
			t.Fatal("case:", a, "result:", f)
		}
		if a.maskFirst == a.maskNext {
			f = IsType(a.maskFirst, a.str)
			if f != a.is {
				t.Fatal("case:", a, "result:", f)
			}
		}
	}
}

func TestIsCSymbol(t *testing.T) {
	type testCase struct {
		str string
		is  bool
	}
	cases := []testCase{
		{"123", false},
		{"_123", true},
	}
	for _, c := range cases {
		if ret := IsCSymbol(c.str); ret != c.is {
			t.Fatal("IsCSymbol:", c.str, "got:", ret)
		}
	}
}

func TestIsXmlSymbol(t *testing.T) {
	type testCase struct {
		str string
		is  bool
	}
	cases := []testCase{
		{"123", false},
		{"_123", true},
	}
	for _, c := range cases {
		if ret := IsXmlSymbol(c.str); ret != c.is {
			t.Fatal("IsXmlSymbol:", c.str, "got:", ret)
		}
	}
}

func TestScanType(t *testing.T) {
	type testCase struct {
		str string
		pos int
	}
	cases := []testCase{
		{"123", 0},
		{"_123", 1},
		{"_xml", -1},
	}
	for _, c := range cases {
		if ret := ScanType(CSYMBOL_FIRST_CHAR, c.str); ret != c.pos {
			t.Fatal("ScanType:", c.str, "got:", ret)
		}
	}
}

func TestScanCSymbol(t *testing.T) {
	type testCase struct {
		str string
		pos int
	}
	cases := []testCase{
		{"123", 0},
		{"_123", -1},
		{"_123 xml", 4},
	}
	for _, c := range cases {
		if ret := ScanCSymbol(c.str); ret != c.pos {
			t.Fatal("ScanCSymbol:", c.str, "got:", ret)
		}
	}
}

func TestScanXmlSymbol(t *testing.T) {
	type testCase struct {
		str string
		pos int
	}
	cases := []testCase{
		{"123", 0},
		{"_123", -1},
		{"_123 xml", 4},
	}
	for _, c := range cases {
		if ret := ScanXmlSymbol(c.str); ret != c.pos {
			t.Fatal("ScanXmlSymbol:", c.str, "got:", ret)
		}
	}
}

func TestSkipXmlSymbol(t *testing.T) {
	type testCase struct {
		str string
		ret string
	}
	cases := []testCase{
		{"123", "123"},
		{"_123", ""},
		{"_123 xml", " xml"},
	}
	for _, c := range cases {
		if ret := SkipXmlSymbol(c.str); ret != c.ret {
			t.Fatal("SkipXmlSymbol:", c.str, "got:", ret)
		}
	}
}

func TestSkipCSymbol(t *testing.T) {
	type testCase struct {
		str string
		ret string
	}
	cases := []testCase{
		{"123", "123"},
		{"_123", ""},
		{"_123 xml", " xml"},
	}
	for _, c := range cases {
		if ret := SkipCSymbol(c.str); ret != c.ret {
			t.Fatal("SkipCSymbol:", c.str, "got:", ret)
		}
	}
}
