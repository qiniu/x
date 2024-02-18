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

package gsh

import (
	"io"
	"os"
	"os/exec"
	"testing"
)

type mockOS struct{}

func (p mockOS) Environ() []string {
	return mockEnv
}

func (p mockOS) ExpandEnv(s string) string {
	return os.Expand(s, func(name string) string {
		return Getenv(mockEnv, name)
	})
}

func (p mockOS) Run(c *exec.Cmd) error {
	if mockRunOut != "" {
		io.WriteString(c.Stdout, mockRunOut)
	}
	return mockRunErr
}

var (
	mockEnv    []string
	mockRunOut string
	mockRunErr error
)

func init() {
	Sys = mockOS{}
}

// -----------------------------------------------------------

func TestBasic(t *testing.T) {
	var app App
	app.initApp()
	err := app.Gop_Exec("ls", "-l")
	check(t, err)
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// -----------------------------------------------------------
