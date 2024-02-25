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
	"os"
	"os/exec"
	"strings"
)

// -----------------------------------------------------------

// Setenv overwrites environments with specified env.
func Setenv__0(ret []string, env map[string]string) []string {
	for k, v := range env {
		nameEq := k + "="
		ret = setenv(ret, nameEq+v, len(nameEq))
	}
	return ret
}

// Setenv overwrites environments with specified (name, val) pair.
func Setenv__1(ret []string, name, val string) []string {
	nameEq := name + "="
	return setenv(ret, nameEq+val, len(nameEq))
}

func setenv(ret []string, pair string, idxVal int) []string {
	nameEq := pair[:idxVal]
	for i, e := range ret {
		if strings.HasPrefix(e, nameEq) {
			ret[i] = pair
			return ret
		}
	}
	return append(ret, pair)
}

// Setenv overwrites environments with specified "name=val" pairs.
func Setenv__2(ret []string, env []string) []string {
	for _, pair := range env {
		pos := strings.IndexByte(pair, '=')
		if pos > 0 {
			ret = setenv(ret, pair, pos+1)
		}
	}
	return ret
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
// To distinguish between an empty value and an unset value, use LookupEnv.
func Getenv(env []string, name string) string {
	nameEq := name + "="
	for _, e := range env {
		if strings.HasPrefix(e, nameEq) {
			return e[len(nameEq):]
		}
	}
	return ""
}

// -----------------------------------------------------------

type OS interface {
	// Environ returns a copy of strings representing the environment,
	// in the form "key=value".
	Environ() []string

	// ExpandEnv replaces ${var} or $var in the string according to the values
	// of the current environment variables. References to undefined
	// variables are replaced by the empty string.
	ExpandEnv(s string) string

	// Getenv retrieves the value of the environment variable named by the key.
	// It returns the value, which will be empty if the variable is not present.
	// To distinguish between an empty value and an unset value, use LookupEnv.
	Getenv(key string) string

	// Run starts the specified command and waits for it to complete.
	Run(c *exec.Cmd) error
}

// -----------------------------------------------------------

type defaultOS struct{}

func (p defaultOS) Environ() []string {
	return os.Environ()
}

func (p defaultOS) ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

func (p defaultOS) Getenv(key string) string {
	return os.Getenv(key)
}

func (p defaultOS) Run(c *exec.Cmd) error {
	return c.Run()
}

var Sys OS = defaultOS{}

// -----------------------------------------------------------
