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

package jsonutil

import (
	"encoding/json"

	"github.com/qiniu/x/byteutil"
	"github.com/qiniu/x/stringutil"
)

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data string, v interface{}) error {
	b := byteutil.Bytes(data)
	return json.Unmarshal(b, v)
}

// ----------------------------------------------------------

// Stringify converts a value into string.
func Stringify(v interface{}) string {
	b, _ := json.Marshal(v)
	return stringutil.String(b)
}

// ----------------------------------------------------------
