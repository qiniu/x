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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	GopPackage = true
)

type basetype interface {
	string | int | bool | float64
}

func toMapAny[T basetype](val map[string]T) map[string]any {
	ret := make(map[string]any, len(val))
	for k, v := range val {
		ret[k] = v
	}
	return ret
}

func toSlice(val []map[string]any) []any {
	ret := make([]any, len(val))
	for i, v := range val {
		ret[i] = v
	}
	return ret
}

func toSliceAny[T basetype](val []map[string]T) []any {
	ret := make([]any, len(val))
	for i, v := range val {
		ret[i] = toMapAny(v)
	}
	return ret
}

func tryToMapAny(val any) (ret map[string]any, ok bool) {
	v := reflect.ValueOf(val)
	return castMapAny(v)
}

func castMapAny(v reflect.Value) (ret map[string]any, ok bool) {
	if v.Kind() != reflect.Map || v.Type().Key() != tyString {
		return
	}
	ret, ok = make(map[string]any, v.Len()), true
	for it := v.MapRange(); it.Next(); {
		key := it.Key().String()
		ret[key] = it.Value().Interface()
	}
	return
}

var (
	tyString = reflect.TypeOf("")
)

// -----------------------------------------------------------------------------

type baseelem interface {
	string
}

type baseslice interface {
	[]string
}

type TySet[T baseelem] []T
type TyAnySet []any

func Set__0[T baseelem](vals ...T) TySet[T] {
	return TySet[T](vals)
}

func Set__1[T []string](v *Var__3[T]) TySet[string] {
	return TySet[string](v.Val())
}

func Set__2(vals ...any) TyAnySet {
	return TyAnySet(vals)
}

// -----------------------------------------------------------------------------

// Case represents a test case.
type Case struct {
	CaseT
}

func nameCtx(name []string) string {
	if name != nil {
		return " (" + strings.Join(name, ".") + ")"
	}
	return ""
}

const (
	Gopo_Gopt_Case_Match = "Gopt_Case_MatchTBase,Gopt_Case_MatchMap,Gopt_Case_MatchSlice,Gopt_Case_MatchBaseSlice,Gopt_Case_MatchSet,Gopt_Case_MatchAnySet,Gopt_Case_MatchAny"
)

func Gopt_Case_MatchTBase[T basetype](t CaseT, expected, got T, name ...string) {
	if expected != got {
		t.Helper()
		t.Fatalf("unmatched value%s - expected: %v, got: %v\n", nameCtx(name), expected, got)
	}
}

func Gopt_Case_MatchMap(t CaseT, expected, got map[string]any, name ...string) {
	t.Helper()
	idx := len(name)
	name = append(name, "")
	for key, ev := range expected {
		name[idx] = key
		Gopt_Case_MatchAny(t, ev, got[key], name...)
	}
}

func Gopt_Case_MatchSlice[T any](t CaseT, expected []T, got []any, name ...string) {
	t.Helper()
	if len(expected) != len(got) {
		t.Fatalf("unmatched slice%s length - expected: %d, got: %d\n", nameCtx(name), len(expected), len(got))
	}
	idx := len(name) - 1
	if idx < 0 {
		idx, name = 0, []string{"$"}
	}
	for i, ev := range expected {
		name[idx] = "[" + strconv.Itoa(i) + "]"
		Gopt_Case_MatchAny(t, ev, got[i], name...)
	}
}

func Gopt_Case_MatchBaseSlice[T baseelem](t CaseT, expected, got []T, name ...string) {
	t.Helper()
	if len(expected) != len(got) {
		t.Fatalf("unmatched slice%s length - expected: %d, got: %d\n", nameCtx(name), len(expected), len(got))
	}
	idx := len(name) - 1
	if idx < 0 {
		idx, name = 0, []string{"$"}
	}
	for i, ev := range expected {
		name[idx] = "[" + strconv.Itoa(i) + "]"
		Gopt_Case_MatchTBase(t, ev, got[i], name...)
	}
}

func Gopt_Case_MatchSet[T baseelem](t CaseT, expected TySet[T], got []T, name ...string) {
	if len(expected) != len(got) {
		t.Fatalf("unmatched set%s length - expected: %d, got: %d\n", nameCtx(name), len(expected), len(got))
	}
	for _, gv := range got {
		if !hasElem(gv, expected) {
			t.Fatalf("unmatched set%s: expected: %v, value %v doesn't exist in it\n", nameCtx(name), expected, gv)
		}
	}
}

func Gopt_Case_MatchAnySet(t CaseT, expected TyAnySet, got any, name ...string) {
	if gv, ok := got.([]any); ok {
		matchAnySet(t, expected, gv)
		return
	}
	vgot := reflect.ValueOf(got)
	if vgot.Kind() != reflect.Slice {
		t.Fatalf("unmatched set%s: expected: %v, got a non slice value: %v\n", nameCtx(name), expected, got)
	}
	for i, n := 0, vgot.Len(); i < n; i++ {
		gv := vgot.Index(i).Interface()
		if !hasAnyElem(gv, expected) {
			t.Fatalf("unmatched set%s: expected: %v, value %v doesn't exist in it\n", nameCtx(name), expected, gv)
		}
	}
}

func matchAnySet(t CaseT, expected TyAnySet, got []any, name ...string) {
	if len(expected) != len(got) {
		t.Fatalf("unmatched set%s length - expected: %d, got: %d\n", nameCtx(name), len(expected), len(got))
	}
	for _, gv := range got {
		if !hasAnyElem(gv, expected) {
			t.Fatalf("unmatched set%s: expected: %v, value %v doesn't exist in it\n", nameCtx(name), expected, gv)
		}
	}
}

func hasElem[T baseelem](v T, expected []T) bool {
	for _, ev := range expected {
		if reflect.DeepEqual(v, ev) {
			return true
		}
	}
	return false
}

func hasAnyElem(v any, expected []any) bool {
	for _, ev := range expected {
		if v == ev {
			return true
		}
	}
	return false
}

func Gopt_Case_MatchAny(t CaseT, expected, got any, name ...string) {
	t.Helper()
retry:
	switch ev := expected.(type) {
	case string:
		switch gv := got.(type) {
		case string:
			Gopt_Case_MatchTBase(t, ev, gv, name...)
			return
		case *Var__0[string]:
			Gopt_Case_MatchTBase(t, ev, gv.Val(), name...)
			return
		}
	case int:
		switch gv := got.(type) {
		case int:
			Gopt_Case_MatchTBase(t, ev, gv, name...)
			return
		case *Var__0[int]:
			Gopt_Case_MatchTBase(t, ev, gv.Val(), name...)
			return
		}
	case bool:
		switch gv := got.(type) {
		case bool:
			Gopt_Case_MatchTBase(t, ev, gv, name...)
			return
		case *Var__0[bool]:
			Gopt_Case_MatchTBase(t, ev, gv.Val(), name...)
			return
		}
	case float64:
		switch gv := got.(type) {
		case float64:
			Gopt_Case_MatchTBase(t, ev, gv, name...)
			return
		case *Var__0[float64]:
			Gopt_Case_MatchTBase(t, ev, gv.Val(), name...)
			return
		}
	case map[string]any:
		switch gv := got.(type) {
		case map[string]any:
			Gopt_Case_MatchMap(t, ev, gv, name...)
			return
		case *Var__1[map[string]any]:
			Gopt_Case_MatchMap(t, ev, gv.Val(), name...)
			return
		default:
			if gv, ok := tryToMapAny(got); ok {
				Gopt_Case_MatchMap(t, ev, gv, name...)
				return
			}
		}
	case []any:
		switch gv := got.(type) {
		case []any:
			Gopt_Case_MatchSlice(t, ev, gv, name...)
			return
		case *Var__2[[]any]:
			Gopt_Case_MatchSlice(t, ev, gv.Val(), name...)
			return
		}
	case []string:
		switch gv := got.(type) {
		case []string:
			Gopt_Case_MatchBaseSlice(t, ev, gv, name...)
			return
		case *Var__3[[]string]:
			Gopt_Case_MatchBaseSlice(t, ev, gv.Val(), name...)
			return
		case []any:
			Gopt_Case_MatchSlice(t, ev, gv, name...)
			return
		}
	case TySet[string]:
		switch gv := got.(type) {
		case []string:
			Gopt_Case_MatchSet(t, ev, gv, name...)
			return
		case *Var__3[[]string]:
			Gopt_Case_MatchSet(t, ev, gv.Val(), name...)
			return
		}
	case TyAnySet:
		switch gv := got.(type) {
		case *Var__2[[]any]:
			Gopt_Case_MatchAnySet(t, ev, gv.Val(), name...)
			return
		default:
			Gopt_Case_MatchAnySet(t, ev, gv, name...)
			return
		}
	case *Var__0[string]:
		switch gv := got.(type) {
		case string:
			ev.Match(t, gv, name...)
			return
		case *Var__0[string]:
			ev.Match(t, gv.Val(), name...)
			return
		case nil:
			ev.MatchNil(t, name...)
			return
		}
	case *Var__0[int]:
		switch gv := got.(type) {
		case int:
			ev.Match(t, gv, name...)
			return
		case *Var__0[int]:
			ev.Match(t, gv.Val(), name...)
			return
		case nil:
			ev.MatchNil(t, name...)
			return
		}
	case *Var__0[bool]:
		switch gv := got.(type) {
		case bool:
			ev.Match(t, gv, name...)
			return
		case *Var__0[bool]:
			ev.Match(t, gv.Val(), name...)
			return
		case nil:
			ev.MatchNil(t, name...)
			return
		}
	case *Var__0[float64]:
		switch gv := got.(type) {
		case float64:
			ev.Match(t, gv, name...)
			return
		case *Var__0[float64]:
			ev.Match(t, gv.Val(), name...)
			return
		case nil:
			ev.MatchNil(t, name...)
			return
		}
	case *Var__1[map[string]any]:
		switch gv := got.(type) {
		case map[string]any:
			ev.Match(t, gv, name...)
			return
		case *Var__1[map[string]any]:
			ev.Match(t, gv.Val(), name...)
			return
		}
	case *Var__2[[]any]:
		switch gv := got.(type) {
		case []any:
			ev.Match(t, gv, name...)
			return
		case *Var__2[[]any]:
			ev.Match(t, gv.Val(), name...)
			return
		}
	case *Var__3[[]string]:
		switch gv := got.(type) {
		case []string:
			ev.Match(t, gv, name...)
			return
		case *Var__3[[]string]:
			ev.Match(t, gv.Val(), name...)
			return
		}

	// fallback types:
	case map[string]string:
		expected = toMapAny(ev)
		goto retry
	case map[string]int:
		expected = toMapAny(ev)
		goto retry
	case map[string]bool:
		expected = toMapAny(ev)
		goto retry
	case map[string]float64:
		expected = toMapAny(ev)
		goto retry

	case []map[string]any:
		expected = toSlice(ev)
		goto retry
	case []map[string]string:
		expected = toSliceAny(ev)
		goto retry

	// other types:
	default:
		if v, ok := tryToMapAny(expected); ok {
			expected = v
			goto retry
		}
		if reflect.DeepEqual(expected, got) {
			return
		}
	}
	t.Fatalf(
		"unmatched%s - expected: %v (%T), got: %v (%T)\n",
		nameCtx(name), expected, expected, got, got,
	)
}

// -----------------------------------------------------------------------------

type Var__0[T basetype] struct {
	val   T
	valid bool
}

func (p *Var__0[T]) check() {
	if !p.valid {
		Fatal("read variable value before initialization")
	}
}

func (p *Var__0[T]) Ok() bool {
	return p.valid
}

func (p *Var__0[T]) String() string {
	p.check()
	return fmt.Sprint(p.val) // TODO(xsw): optimization
}

func (p *Var__0[T]) Val() T {
	p.check()
	return p.val
}

func (p *Var__0[T]) MarshalJSON() ([]byte, error) {
	p.check()
	return json.Marshal(p.val)
}

func (p *Var__0[T]) UnmarshalJSON(data []byte) error {
	p.valid = true
	return json.Unmarshal(data, &p.val)
}

func (p *Var__0[T]) Equal(t CaseT, v T) bool {
	p.check()
	return p.val == v
}

func (p *Var__0[T]) Match(t CaseT, v T, name ...string) {
	if !p.valid {
		p.val, p.valid = v, true
		return
	}
	t.Helper()
	Gopt_Case_MatchTBase(t, p.val, v, name...)
}

func (p *Var__0[T]) MatchNil(t CaseT, name ...string) {
	if p.valid {
		t.Helper()
		t.Fatalf("unmatched%s - expected: nil, got: %v\n", nameCtx(name), p.val)
	}
}

// -----------------------------------------------------------------------------

type Var__1[T map[string]any] struct {
	val T
}

func (p *Var__1[T]) check() {
	if p.val == nil {
		Fatal("read variable value before initialization")
	}
}

func (p *Var__1[T]) Ok() bool {
	return p.val != nil
}

func (p *Var__1[T]) Val() T {
	p.check()
	return p.val
}

func (p *Var__1[T]) MarshalJSON() ([]byte, error) {
	p.check()
	return json.Marshal(p.val)
}

func (p *Var__1[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.val)
}

func (p *Var__1[T]) Match(t CaseT, v T, name ...string) {
	if p.val == nil {
		p.val = v
		return
	}
	t.Helper()
	Gopt_Case_MatchMap(t, p.val, v, name...)
}

// -----------------------------------------------------------------------------

type Var__2[T []any] struct {
	val   T
	valid bool
}

func (p *Var__2[T]) check() {
	if !p.valid {
		Fatal("read variable value before initialization")
	}
}

func (p *Var__2[T]) Ok() bool {
	return p.valid
}

func (p *Var__2[T]) Val() T {
	p.check()
	return p.val
}

func (p *Var__2[T]) MarshalJSON() ([]byte, error) {
	p.check()
	return json.Marshal(p.val)
}

func (p *Var__2[T]) UnmarshalJSON(data []byte) error {
	p.valid = true
	return json.Unmarshal(data, &p.val)
}

func (p *Var__2[T]) Match(t CaseT, v T, name ...string) {
	if p.val == nil {
		p.val, p.valid = v, true
		return
	}
	t.Helper()
	Gopt_Case_MatchSlice(t, p.val, v, name...)
}

// -----------------------------------------------------------------------------

type Var__3[T baseslice] struct {
	val   T
	valid bool
}

func (p *Var__3[T]) check() {
	if !p.valid {
		Fatal("read variable value before initialization")
	}
}

func (p *Var__3[T]) Ok() bool {
	return p.valid
}

func (p *Var__3[T]) Val() T {
	p.check()
	return p.val
}

func (p *Var__3[T]) MarshalJSON() ([]byte, error) {
	p.check()
	return json.Marshal(p.val)
}

func (p *Var__3[T]) UnmarshalJSON(data []byte) error {
	p.valid = true
	return json.Unmarshal(data, &p.val)
}

func (p *Var__3[T]) Match(t CaseT, v T, name ...string) {
	if p.val == nil {
		p.val, p.valid = v, true
		return
	}
	t.Helper()
	Gopt_Case_MatchBaseSlice(t, p.val, v, name...)
}

// -----------------------------------------------------------------------------

func Gopx_Var_Cast__0[T basetype]() *Var__0[T] {
	return new(Var__0[T])
}

func Gopx_Var_Cast__1[T map[string]any]() *Var__1[T] {
	return new(Var__1[T])
}

func Gopx_Var_Cast__2[T []any]() *Var__2[T] {
	return new(Var__2[T])
}

func Gopx_Var_Cast__3[T []string]() *Var__3[T] {
	return new(Var__3[T])
}

// -----------------------------------------------------------------------------
