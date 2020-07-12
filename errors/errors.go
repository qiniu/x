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

// Package errors provides errors stack tracking utilities.
package errors

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
)

// --------------------------------------------------------------------

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(msg string) error {
	return errors.New(msg)
}

// Err returns the cause error.
func Err(err error) error {
	if e, ok := err.(*Frame); ok {
		return Err(e.Err)
	}
	return err
}

// --------------------------------------------------------------------

// Frame represents an error frame.
type Frame struct {
	Err  error
	Pkg  string
	Func string
	Args []interface{}
	Code string
	File string
	Line int
}

// NewFrame creates a new error frame.
func NewFrame(err error, code, file string, line int, pkg, fn string, args ...interface{}) *Frame {
	return &Frame{Err: err, Pkg: pkg, Func: fn, Args: args, Code: code, File: file, Line: line}
}

func (p *Frame) Error() string {
	return string(errorDetail(make([]byte, 0, 32), p))
}

func errorDetail(b []byte, p *Frame) []byte {
	if f, ok := p.Err.(*Frame); ok {
		b = errorDetail(b, f)
	} else {
		b = append(b, p.Err.Error()...)
		b = append(b, "\n\n===> errors stack:\n"...)
	}
	b = append(b, p.Pkg...)
	b = append(b, '.')
	b = append(b, p.Func...)
	b = append(b, '(')
	b = argsDetail(b, p.Args)
	b = append(b, ")\n\t"...)
	b = append(b, p.File...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(p.Line), 10)
	b = append(b, ' ')
	b = append(b, p.Code...)
	b = append(b, '\n')
	return b
}

func argsDetail(b []byte, args []interface{}) []byte {
	nlast := len(args) - 1
	for i, arg := range args {
		b = appendValue(b, arg)
		if i != nlast {
			b = append(b, ',', ' ')
		}
	}
	return b
}

func appendValue(b []byte, arg interface{}) []byte {
	if arg == nil {
		return append(b, "nil"...)
	}
	v := reflect.ValueOf(arg)
	kind := v.Kind()
	if kind >= reflect.Bool && kind <= reflect.Complex128 {
		return append(b, fmt.Sprint(arg)...)
	}
	if kind == reflect.String {
		val := arg.(string)
		if len(val) > 16 {
			val = val[:16] + "..."
		}
		return strconv.AppendQuote(b, val)
	}
	if kind == reflect.Array {
		return append(b, "Array"...)
	}
	if kind == reflect.Struct {
		return append(b, "Struct"...)
	}
	val := v.Pointer()
	b = append(b, '0', 'x')
	return strconv.AppendInt(b, int64(val), 16)
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (p *Frame) Unwrap() error {
	return p.Err
}

// Format is required by fmt.Formatter
func (p *Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		io.WriteString(s, p.Error())
	case 's':
		io.WriteString(s, Err(p.Err).Error())
	case 'q':
		fmt.Fprintf(s, "%q", Err(p.Err).Error())
	}
}

// --------------------------------------------------------------------

// CallDetail print a function call shortly.
func CallDetail(msg []byte, fn interface{}, args ...interface{}) []byte {
	f := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if f != nil {
		msg = append(msg, f.Name()...)
		msg = append(msg, '(')
		msg = argsDetail(msg, args)
		msg = append(msg, ')')
	}
	return msg
}

// --------------------------------------------------------------------

// ErrorInfo is provided for backward compatibility
type ErrorInfo = Frame

// Detail is provided for backward compatibility
func (p *ErrorInfo) Detail(err error) *ErrorInfo {
	p.Code = err.Error()
	return p
}

// NestedObject is provided for backward compatibility
func (p *ErrorInfo) NestedObject() interface{} {
	return p.Err
}

// ErrorDetail is provided for backward compatibility
func (p *ErrorInfo) ErrorDetail() string {
	return p.Error()
}

// AppendErrorDetail is provided for backward compatibility
func (p *ErrorInfo) AppendErrorDetail(b []byte) []byte {
	return errorDetail(b, p)
}

// SummaryErr is provided for backward compatibility
func (p *ErrorInfo) SummaryErr() error {
	return p.Err
}

// Info is provided for backward compatibility
func Info(err error, cmd ...interface{}) *ErrorInfo {
	return &Frame{Err: err, Args: cmd}
}

// InfoEx is provided for backward compatibility
func InfoEx(calldepth int, err error, cmd ...interface{}) *ErrorInfo {
	return &Frame{Err: err, Args: cmd}
}

// Detail is provided for backward compatibility
func Detail(err error) string {
	return err.Error()
}

// --------------------------------------------------------------------
