/*
 Copyright 2023 Qiniu Limited (qiniu.com)

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

package plugins

import (
	"net/http"
	"path"
)

type Handler = func(w http.ResponseWriter, r *http.Request, next http.Handler)

func New(h http.Handler, plugins ...interface{}) http.Handler {
	n := len(plugins)
	exts := make(map[string]Handler, n/2)
	for i := 0; i < n; i += 2 {
		ext := plugins[i].(string)
		fn := plugins[i+1].(Handler)
		exts[ext] = fn
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := path.Ext(r.URL.Path)
		if fn, ok := exts[ext]; ok {
			fn(w, r, h)
			return
		}
		h.ServeHTTP(w, r)
	})
}
