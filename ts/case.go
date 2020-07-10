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

// Package ts provides Go packages testing utilities.
package ts

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/qiniu/x/errors"
)

// ----------------------------------------------------------------------------

// Testing represents a testing object.
type Testing struct {
	t *testing.T
}

// New creates a testing object.
func New(t *testing.T) *Testing {
	return &Testing{t: t}
}

// Case creates a test case.
func (p *Testing) Case(name string) *TestCase {
	return &TestCase{name: name, t: p.t}
}

// Call creates a test case, and then calls a function.
func (p *Testing) Call(fn interface{}, args ...interface{}) *TestCase {
	return p.Case("").Call(fn, args...)
}

// ----------------------------------------------------------------------------

// TestCase represents a test case.
type TestCase struct {
	t    *testing.T
	name string
	msg  []byte
	out  []reflect.Value
	with reflect.Value
}

// Call calls a function.
func (p *TestCase) Call(fn interface{}, args ...interface{}) *TestCase {
	msg := make([]byte, 0, 16)
	if p.name != "" {
		msg = append(msg, p.name...)
		msg = append(msg, ' ')
	}
	f := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if f != nil {
		msg = append(msg, f.Name()...)
		msg = append(msg, '(')
		msg = errors.ArgsDetail(msg, args)
		p.msg = append(msg, ')')
	}
	vfn := reflect.ValueOf(fn)
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	p.out = vfn.Call(in)
	if len(p.out) > 0 {
		p.with = p.out[0]
	}
	return p
}

// With set current output value to check.
func (p *TestCase) With(i int) *TestCase {
	p.with = p.out[i]
	return p
}

// Equal checks current output value.
func (p *TestCase) Equal(v interface{}) *TestCase {
	p.assertEq(p.with.Interface(), v)
	return p
}

// PropEqual checks property of current output value.
func (p *TestCase) PropEqual(prop string, v interface{}) *TestCase {
	o := PropVal(p.with, prop)
	p.assertEq(o.Interface(), v)
	return p
}

func (p *TestCase) assertEq(a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		p.t.Fatalf("%s:\nassertEq failed - %v, expected: %v\n", string(p.msg), a, b)
	}
}

// PropVal returns property value of an object.
func PropVal(o reflect.Value, prop string) reflect.Value {
	if o.Kind() == reflect.Struct {
		if ret := o.FieldByName(prop); ret.IsValid() {
			return ret
		}
	}
	if m := o.MethodByName(strings.Title(prop)); m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) != 1 {
			panic("invalid PropVal: " + prop)
		}
		return out[0]
	}
	return reflect.Value{}
}

// ----------------------------------------------------------------------------
