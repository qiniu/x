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
	"unicode"
	"unicode/utf8"
)

// Concat concatenates parts of a string together.
func Concat(parts ...string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	n := 0
	for _, part := range parts {
		n += len(part)
	}
	b := make([]byte, 0, n)
	for _, part := range parts {
		b = append(b, part...)
	}
	return String(b)
}

// Capitalize returns a copy of the string str with the first letter mapped to
// its upper case.
func Capitalize(str string) string {
	c, nc := utf8.DecodeRuneInString(str)
	if unicode.IsUpper(c) || c == utf8.RuneError {
		return str
	}
	ret := make([]byte, len(str))
	nr := utf8.EncodeRune(ret, unicode.ToUpper(c))
	ret = append(ret[:nr], str[nc:]...)
	return String(ret)
}

// Contains checks if the classVal is present in the classAttr string, which
// is a space-separated list of class names.
func Contains(classAttr, classVal string) bool {
	n := len(classAttr)
	m := len(classVal)
	if m == 0 {
		return false
	}

	i := 0
	for i < n {
		// Skip whitespace
		for i < n && classAttr[i] == ' ' {
			i++
		}
		// Mark the start of the current token
		start := i
		// Advance to the end of the token
		for i < n && classAttr[i] != ' ' {
			i++
		}
		// Exact match against classCheck
		if i-start == m && classAttr[start:i] == classVal {
			return true
		}
	}
	return false
}
