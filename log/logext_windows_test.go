//go:build windows
// +build windows

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
