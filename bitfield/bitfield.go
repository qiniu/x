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

// Package bitfield provides bit-field access helpers.
//
// Bit numbering: bit k of the memory at base is bit (k % 8) of byte (k / 8),
// where bit 0 of a byte is its least significant bit (little-endian bit
// allocation, as on x86-64 and AArch64).
//
// Unsigned, Signed and Set require 1 <= width <= 64 and touch only the bytes
// that contain at least one bit of the range [off, off+width). A range of up
// to 64 bits can span 9 bytes when it is not byte-aligned; that case is
// handled explicitly. The helpers are not atomic.
package bitfield

import (
	"unsafe"
)

// Unsigned returns bits off .. off+width-1 of the memory at base,
// zero-extended into the low bits of the result.
func Unsigned(base unsafe.Pointer, off, width uintptr) uint64 {
	s := unsafe.Slice((*uint8)(base), (off+width+7)/8)
	var r uint64
	for i := uintptr(0); i < width; i++ {
		if s[(off+i)/8]&(uint8(1)<<((off+i)%8)) != 0 {
			r |= uint64(1) << i
		}
	}
	return r
}

// Signed returns the same bits as Unsigned, sign-extended
// from bit width-1.
func Signed(base unsafe.Pointer, off, width uintptr) int64 {
	r := Unsigned(base, off, width)
	shift := 64 - width
	return int64(r<<shift) >> shift
}

// Set replaces bits off .. off+width-1 of the memory at base with
// the low width bits of v. All other bits are left unchanged.
func Set(base unsafe.Pointer, off, width uintptr, v uint64) {
	s := unsafe.Slice((*uint8)(base), (off+width+7)/8)
	for i := uintptr(0); i < width; i++ {
		idx := (off + i) / 8
		mask := uint8(1) << ((off + i) % 8)
		if v>>i&1 != 0 {
			s[idx] |= mask
		} else {
			s[idx] &^= mask
		}
	}
}
