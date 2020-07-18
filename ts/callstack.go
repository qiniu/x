/*
 Copyright 2020 Qiniu Limited (qiniu.com)

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

package ts

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// --------------------------------------------------------------------

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

func (f Frame) pc() uintptr { return uintptr(f) - 1 }

func (f Frame) fileLine() (string, int) {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown", 0
	}
	return fn.FileLine(f.pc())
}

// name returns the name of this function, if known.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format formats the frame according to the fmt.Formatter interface.
//
//    %n    <funcname>
//    %v    <file>:<line>
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+v   equivalent to <funcname>\n\t<file>:<line>
//
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		file, line := f.fileLine()
		if s.Flag('+') {
			io.WriteString(s, f.name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, file)
		} else {
			io.WriteString(s, path.Base(file))
		}
		io.WriteString(s, ":")
		io.WriteString(s, strconv.Itoa(line))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	default:
		panic("Frame.Format: unsupport verb - " + string(verb))
	}
}

// --------------------------------------------------------------------

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			fmt.Fprintf(st, "\ngoroutine %d [running]:", len(*s))
			fallthrough
		default:
			for _, pc := range *s {
				f := Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func callers(skip int) *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// funcname removes the path prefix component of a function's name reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// --------------------------------------------------------------------
