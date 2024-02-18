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
	"strings"
)

// Setenv overwrites environments with specified env.
func Setenv__0(ret []string, env map[string]string) []string {
	for k, v := range env {
		ret = Setenv__1(ret, k, v)
	}
	return ret
}

// Setenv overwrites environments with specified (name, val) pair.
func Setenv__1(ret []string, name, val string) []string {
	name += "="
	for i, e := range ret {
		if strings.HasPrefix(e, name) {
			ret[i] = name + val
			return ret
		}
	}
	return append(ret, name+val)
}
