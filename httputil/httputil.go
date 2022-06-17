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

package httputil

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

// ----------------------------------------------------------

// Reply replies a http request with a json response.
func Reply(w http.ResponseWriter, code int, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
}

// ReplyWith replies a http request with a bodyType response.
func ReplyWith(w http.ResponseWriter, code int, bodyType string, msg []byte) {
	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", bodyType)
	w.WriteHeader(code)
	w.Write(msg)
}

// ReplyWithStream replies a http request with a streaming response.
func ReplyWithStream(w http.ResponseWriter, code int, bodyType string, body io.Reader, bytes int64) {
	h := w.Header()
	h.Set("Content-Length", strconv.FormatInt(bytes, 10))
	h.Set("Content-Type", bodyType)
	w.WriteHeader(code)
	// We don't use io.CopyN: if you need, call io.LimitReader(body, bytes) by yourself
	written, err := io.Copy(w, body)
	if err != nil || written != bytes {
		log.Printf("ReplyWithStream (bytes=%v): written=%v, err=%v\n", bytes, written, err)
	}
}

// ----------------------------------------------------------
