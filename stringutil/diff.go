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

// Diff calculates difference of two sorted string (in ASC order).
func Diff(new, old []string) (add, del []string) {
	i, j := 0, 0
	for i != len(new) && j != len(old) {
		if new[i] < old[j] {
			add = append(add, new[i])
			i++
		} else if old[j] < new[i] {
			del = append(del, old[j])
			j++
		} else {
			i++
			j++
		}
	}
	if i < len(new) {
		add = append(add, new[i:]...)
	} else {
		del = append(del, old[j:]...)
	}
	return
}
