//go:build go1.21
// +build go1.21

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

package byteutil

import (
	"unsafe"
)

// Bytes returns a byte slice whose underlying data is s.
func Bytes(s string) []byte {
	// Although unsafe.SliceData/String was introduced in go1.20, but
	// the go version in go.mod is 1.18 so we cannot use them.
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
