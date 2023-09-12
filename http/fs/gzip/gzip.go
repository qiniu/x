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

package gzip

import (
	"compress/gzip"
	"net/http"

	xfs "github.com/qiniu/x/http/fs"
)

func Open(fs http.FileSystem, name string) (file http.File, err error) {
	file, err = fs.Open(name)
	if err != nil {
		return
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	return xfs.SequenceFile(name, gr), nil
}
