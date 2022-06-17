/*
 Copyright 2022 Qiniu Limited (qiniu.com)

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

package x_test

import (
	"fmt"
	"strings"
	"syscall"
	"testing"

	"github.com/qiniu/x/errors"
)

// --------------------------------------------------------------------

func Foo() error {
	return syscall.ENOENT
}

func TestNewWith(t *testing.T) {
	err := Foo()
	if err != nil {
		err = errors.NewWith(err, `Foo()`, -2, "x_test.Foo")
		if strings.Index(err.Error(), "===> errors stack:\nx_test.Foo()\n\t") < 0 {
			t.Fatal(err)
		}
	}
}

func TestErrList(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	e1 := errors.NewWith(err1, ``, -2, "errors.New", "error 1")
	e2 := errors.NewWith(err2, ``, -2, "errors.New", "error 2")
	e := errors.List{e1, e2}
	if v := fmt.Sprintf("%q", e); v != fmt.Sprintf("%q", e.Error()) {
		t.Fatal(v)
	}
	if v := fmt.Sprintf("%v", e); v != fmt.Sprintf("%v", e.Error()) {
		t.Fatal(v)
	}
	if v := fmt.Sprintf("%s", e); v != `error 1
error 2` {
		t.Fatal(v)
	}
}

// --------------------------------------------------------------------
