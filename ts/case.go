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
// See https://github.com/qiniu/x/wiki/How-to-write-a-TestCase for details.
package ts

import (
	"reflect"
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

// New creates a test case.
func (p *Testing) New(name string) *TestCase {
	return &TestCase{name: name, t: p.t}
}

// Call creates a test case, and then calls a function.
func (p *Testing) Call(fn interface{}, args ...interface{}) *TestCase {
	return p.New("").Call(fn, args...)
}

// Case creates a test case and sets its output parameters.
func (p *Testing) Case(name string, result ...interface{}) *TestCase {
	return p.New(name).Init(result...)
}

// ----------------------------------------------------------------------------

// TestCase represents a test case.
type TestCase struct {
	t    *testing.T
	name string
	msg  []byte
	rcov interface{}
	cstk *stack
	out  []reflect.Value
	idx  int
}

func (p *TestCase) newMsg() []byte {
	msg := make([]byte, 0, 16)
	if p.name != "" {
		msg = append(msg, p.name...)
		msg = append(msg, ' ')
	}
	return msg
}

// Init sets output parameters.
func (p *TestCase) Init(result ...interface{}) *TestCase {
	out := make([]reflect.Value, len(result))
	for i, ret := range result {
		out[i] = reflect.ValueOf(ret)
	}
	p.msg = p.newMsg()
	p.rcov = nil
	p.out = out
	p.idx = 0
	return p
}

// Call calls a function.
func (p *TestCase) Call(fn interface{}, args ...interface{}) (e *TestCase) {
	e = p
	e.msg = errors.CallDetail(e.newMsg(), fn, args...)
	defer func() {
		if e.rcov = recover(); e.rcov != nil {
			e.cstk = callers(3)
		}
	}()
	e.rcov = nil
	e.out = reflect.ValueOf(fn).Call(makeArgs(args))
	e.idx = 0
	return
}

func makeArgs(args []interface{}) []reflect.Value {
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	return in
}

// Next sets current output value to next output parameter.
func (p *TestCase) Next() *TestCase {
	p.idx++
	return p
}

// With sets current output value to check.
func (p *TestCase) With(i int) *TestCase {
	p.idx = i
	return p
}

// Panic checks if function call panics or not. Panic(v) means
// function call panics with `v`. If v == nil, it means we don't
// care any detail information about panic.
func (p *TestCase) Panic(panicMsg ...interface{}) *TestCase {
	if panicMsg == nil {
		p.assertNotPanic()
	} else {
		assertPanic(p.t, p.msg, p.rcov, panicMsg[0])
	}
	return p
}

func assertPanic(t *testing.T, msg []byte, rcov interface{}, panicMsg interface{}) {
	if rcov == nil {
		t.Fatalf("%s:\nPanic checks: no panic, expected: panic\n", string(msg))
	}
	if panicMsg != nil {
		if !reflect.DeepEqual(rcov, panicMsg) {
			t.Fatalf("%s:\nPanic checks: %v, expected: %v\n", string(msg), rcov, panicMsg)
		}
	}
}

func (p *TestCase) assertNotPanic() {
	if p.rcov != nil {
		p.t.Fatalf("panic: %v\n%+v\n", p.rcov, p.cstk)
	}
}

// Equal checks current output value.
func (p *TestCase) Equal(v interface{}) *TestCase {
	p.assertNotPanic()
	p.assertEq(p.out[p.idx].Interface(), v)
	return p
}

// PropEqual checks property of current output value.
func (p *TestCase) PropEqual(prop string, v interface{}) *TestCase {
	p.assertNotPanic()
	o := PropVal(p.out[p.idx], prop)
	p.assertEq(o.Interface(), v)
	return p
}

func (p *TestCase) assertEq(a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		p.t.Fatalf("%s:\nassertEq failed: %v, expected: %v\n", string(p.msg), a, b)
	}
}

// PropVal returns property value of an object.
func PropVal(o reflect.Value, prop string) reflect.Value {
start:
	switch o.Kind() {
	case reflect.Struct:
		if ret := o.FieldByName(prop); ret.IsValid() {
			return ret
		}
	case reflect.Interface:
		o = o.Elem()
		goto start
	}
	if m := o.MethodByName(strings.Title(prop)); m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) != 1 {
			panic("invalid PropVal: " + prop)
		}
		return out[0]
	}
	panic(o.Type().String() + " object hasn't property: " + prop)
}

// ----------------------------------------------------------------------------
