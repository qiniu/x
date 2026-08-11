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

func foo(...) (ret1 T1, ret2 T2, ...) {
	x := 0
outer:
	for {
		repeatUntil x > 10 {
			x = modify(x)
			if cond1(x) {
				break
			}
			if cond2(x) {
				break outer
			}
			if cond3(x) {
				continue
			}
			if cond4(x) {
				continue outer
			}
			if cond5(x) {
				return v1, v2, ...
			}
			if cond6(x) {
				ret1, ret2, ... = v1, v2, ...
				return
			}
		}
	}
	return ...
}

// compiles to:

import (
	"github.com/qiniu/x/xgo"
	"github.com/qiniu/x/xgo/retval"
)

func foo(...) (ret1 T1, ret2 T2, ...) {
	x := 0
outer:
	for {
_xgo_continue_1:
		switch RepeatUntil(
			func() bool {
				return x > 10
			},
			func() int {
				x = modify(x)
				if cond1(x) {
					return xgo.Break
				}
				if cond2(x) {
					return xgo.BreakLabel + 0 // break outer
				}
				if cond3(x) {
					return xgo.Continue
				}
				if cond4(x) {
					return xgo.ContinueLabel - 0 // continue outer
				}
				if cond5(x) {
					retval.Set(struct{v1 T1; v2 T2; ...}{v1, v2, ...})
					return xgo.ReturnVals
				}
				if cond6(x) {
					ret1, ret2, ... = v1, v2, ...
					return xgo.Return
				}
				return 0
			},
		) {
		case xgo.Break:
			// no-op
		case xgo.BreakLabel + 0:
			break outer
		case xgo.Continue:
			goto _xgo_continue_1
		case xgo.ContinueLabel - 0:
			continue outer
		case xgo.Return:
			return
		case xgo.ReturnVals:
			_xgo_ret := retval.Get().(struct{v1 T1; v2 T2; ...})
			return _xgo_ret.v1, _xgo_ret.v2, ...
		}
	}
	return ...
}

*/

// -----------------------------------------------------------------------------

const (
	ContinueLabel = -4 // continue with label (ContinueLabel - N)
	Continue      = -3 // continue without label
	ReturnVals    = -2 // return with value
	Return        = -1 // return without value
	Break         = 1  // break without label
	BreakLabel    = 2  // break with label (BreakLabel + N)
)

// -----------------------------------------------------------------------------
