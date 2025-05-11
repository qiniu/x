/*
 Copyright 2025 Qiniu Limited (qiniu.com)

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

package rfs

import (
	"context"
	"errors"
	"io/fs"
)

var (
	// ErrChangeMark is returned when the change mark is invalid.
	ErrChangeMark = errors.New("invalid change mark")

	// ErrNoMoreChanges is returned when there are no more changes to fetch.
	ErrNoMoreChanges = errors.New("no more changes")
)

// Change represents a change in the file system.
type Change struct {
	Name    string
	OldName string
	Info    fs.FileInfo // nil if deleted
}

// Changes is used to fetch changes from a remote source.
type Changes interface {
	// if markChg != "", it only walks changed files after markChg.
	// if can't understand this markChg, it should return ErrChangeMark.
	// returns ErrNoMoreChanges if no more changes (but chgs still may not empty).
	Changes(ctx context.Context, markChg string) (chgs []Change, markChgNext string, err error)
}
