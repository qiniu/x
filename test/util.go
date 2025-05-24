/*
 * Copyright (c) 2024 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"log"
	"os"
	"testing"
)

// -----------------------------------------------------------------------------

// Fatal is a wrapper for log.Panicln.
func Fatal(v ...any) {
	log.Panicln(v...)
}

// Fatalf is a wrapper for log.Panicf.
func Fatalf(format string, v ...any) {
	log.Panicf(format, v...)
}

// -----------------------------------------------------------------------------

// Diff compares the dst and src byte slices.
// If they are different, it writes the dst to the outfile and logs the differences.
func Diff(t *testing.T, outfile string, dst, src []byte) bool {
	line := 1
	offs := 0 // line offset
	for i := 0; i < len(dst) && i < len(src); i++ {
		d := dst[i]
		s := src[i]
		if d != s {
			os.WriteFile(outfile, dst, 0644)
			t.Errorf("dst:%d: %s\n", line, dst[offs:])
			t.Errorf("src:%d: %s\n", line, src[offs:])
			return true
		}
		if s == '\n' {
			line++
			offs = i + 1
		}
	}
	if len(dst) != len(src) {
		os.WriteFile(outfile, dst, 0644)
		t.Errorf("len(dst) = %d, len(src) = %d\ndst = %q\nsrc = %q", len(dst), len(src), dst, src)
		return true
	}
	return false
}

// -----------------------------------------------------------------------------
