/*
 * Copyright (c) 2024 The GoPlus Authors (goplus.org). All rights reserved.
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

package logt

import (
	"log"
	"time"
)

type T struct {
	name    string
	fail    bool
	skipped bool
}

func New() *T {
	return &T{}
}

func (p *T) Name() string {
	return p.name
}

// Fail marks the function as having failed but continues execution.
func (p *T) Fail() {
	p.fail = true
}

// Failed reports whether the function has failed.
func (p *T) Failed() bool {
	return p.fail
}

// FailNow marks the function as having failed and stops its execution
// by calling runtime.Goexit (which then runs all deferred calls in the
// current goroutine).
// Execution will continue at the next test or benchmark.
// FailNow must be called from the goroutine running the
// test or benchmark function, not from other goroutines
// created during the test. Calling FailNow does not stop
// those other goroutines.
func (p *T) FailNow() {
	p.fail = true
	panic("todo")
}

// Log formats its arguments using default formatting, analogous to Println,
// and records the text in the error log. For tests, the text will be printed only if
// the test fails or the -test.v flag is set. For benchmarks, the text is always
// printed to avoid having performance depend on the value of the -test.v flag.
func (p *T) Log(args ...any) {
	log.Println(args...)
}

// Logf formats its arguments according to the format, analogous to Printf, and
// records the text in the error log. A final newline is added if not provided. For
// tests, the text will be printed only if the test fails or the -test.v flag is
// set. For benchmarks, the text is always printed to avoid having performance
// depend on the value of the -test.v flag.
func (p *T) Logf(format string, args ...any) {
	log.Printf(format, args...)
}

// Errorln is equivalent to Log followed by Fail.
func (p *T) Errorln(args ...any) {
	log.Println(args...)
	p.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (p *T) Errorf(format string, args ...any) {
	log.Printf(format, args...)
	p.Fail()
}

// Fatal is equivalent to Log followed by FailNow.
func (p *T) Fatal(args ...any) {
	log.Panicln(args...)
}

// Fatalf is equivalent to Logf followed by FailNow.
func (p *T) Fatalf(format string, args ...any) {
	log.Panicf(format, args...)
}

// Skip is equivalent to Log followed by SkipNow.
func (p *T) Skip(args ...any) {
	log.Println(args...)
	p.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (p *T) Skipf(format string, args ...any) {
	log.Printf(format, args...)
	p.SkipNow()
}

// SkipNow marks the test as having been skipped and stops its execution
// by calling runtime.Goexit.
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// Execution will continue at the next test or benchmark. See also FailNow.
// SkipNow must be called from the goroutine running the test, not from
// other goroutines created during the test. Calling SkipNow does not stop
// those other goroutines.
func (p *T) SkipNow() {
	p.skipped = true
}

// Skipped reports whether the test was skipped.
func (p *T) Skipped() bool {
	return p.skipped
}

// Helper marks the calling function as a test helper function.
// When printing file and line information, that function will be skipped.
// Helper may be called simultaneously from multiple goroutines.
func (p *T) Helper() {
}

// Cleanup registers a function to be called when the test (or subtest) and all its
// subtests complete. Cleanup functions will be called in last added,
// first called order.
func (p *T) Cleanup(f func()) {
	// TODO:
}

// TempDir returns a temporary directory for the test to use.
// The directory is automatically removed by Cleanup when the test and
// all its subtests complete.
// Each subsequent call to t.TempDir returns a unique directory;
// if the directory creation fails, TempDir terminates the test by calling Fatal.
func (p *T) TempDir() string {
	panic("todo")
}

// Run runs f as a subtest of t called name.
//
// Run may be called simultaneously from multiple goroutines, but all such calls
// must return before the outer test function for t returns.
func (p *T) Run(name string, f func()) bool {
	p.name = name
	f()
	return true
}

// Deadline reports the time at which the test binary will have
// exceeded the timeout specified by the -timeout flag.
//
// The ok result is false if the -timeout flag indicates “no timeout” (0).
func (p *T) Deadline() (deadline time.Time, ok bool) {
	panic("todo")
}
