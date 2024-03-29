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
	"testing"
)

func Test(t *testing.T) {
	var ret struct {
		ID string `json:"id"`
	}
	err := Unmarshal(`{"id": "123"}`, &ret)
	if err != nil {
		t.Fatal("Unmarshal failed:", err)
	}
	if ret.ID != "123" {
		t.Fatal("Unmarshal uncorrect:", ret.ID)
	}
	if v := Stringify(ret); v != `{"id":"123"}` {
		t.Fatal("Stringify:", v)
	}
}
