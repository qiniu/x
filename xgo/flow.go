/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
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

package xgo

/*

func RepeatUntil(__xgo_autoclosure_cond func() bool, body func() int) int {
	for !__xgo_autoclosure_cond() {
		if ret := body(); ret != 0 {
			return ret
		}
	}
	return 0
}

// example:

func foo(...) (T1, T2, ...) {
	x := 0
	repeatUntil x > 10 {
		if cond1(x) {
			break
		}
		if cond2(x) {
			continue
		}
		if cond3(x) {
			return v1, v2, ...
		}
		x = modify(x)
	}
	return ...
}

// compiles to:

func foo(...) (T1, T2, ...) {
	x := 0
lzContinue:
	switch RepeatUntil(
		func() bool {
			return x > 10
		},
		func() int {
			if cond1(x) {
				return Break
			}
			if cond2(x) {
				return Continue
			}
			if cond3(x) {
				xgo.SetRetVal(struct{v1 T1; v2 T2; ...}{v1, v2, ...})
				return Return
			}
			x = modify(x)
			return 0
		},
	) {
	case Break:
		// no-op
	case Continue:
		goto lzContinue
	case Return:
		_xgo_ret := xgo.RetVal().(struct{v1 T1; v2 T2; ...})
		return _xgo_ret.v1, _xgo_ret.v2, ...
	}
	return ...
}

*/

// -----------------------------------------------------------------------------
// gls = goroutine local storage

func _gls_get_retval() any {
	// TODO(xsw):
	return nil
}

func _gls_set_retval(v any) {
	// TODO(xsw):
}

// -----------------------------------------------------------------------------

const (
	Return   = -1
	Break    = 1
	Continue = 2
)

// SetRetVal sets the return value of the current goroutine.
func SetRetVal(v any) {
	_gls_set_retval(v)
}

// RetVal returns the return value of the current goroutine.
func RetVal() any {
	return _gls_get_retval()
}

// -----------------------------------------------------------------------------
