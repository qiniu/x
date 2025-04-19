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

package fs_test

import (
	"net/http"
	"testing"

	"github.com/qiniu/x/http/fs"
	"github.com/qiniu/x/http/fs/filter"
)

func TestLocalCheck(t *testing.T) {
	fsys := http.Dir("/")
	if dir, ok := fs.LocalCheck(fsys); !ok || dir != "/" {
		t.Fatal("fs.LocalCheck(http.Dir):", dir, ok)
	}
	selFs := filter.Select(fsys, "*.txt")
	if dir, ok := fs.LocalCheck(selFs); !ok || dir != "/" {
		t.Fatal("fs.LocalCheck(filterFS):", dir, ok)
	}
	if _, ok := fs.LocalCheck(fs.Root()); ok {
		t.Fatal("fs.LocalCheck(root):", ok)
	}
}
