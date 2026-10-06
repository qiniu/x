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

// Package bitfield provides bit-field access helpers for llcppg generated
// bindings.
//
// The code is the same for every C/C++ package: it only depends on a bit
// offset and a bit width, never on a particular struct.
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
	first := off >> 3
	n := ((off + width - 1) >> 3) - first + 1 // bytes touched, 1..9
	shift := off & 7

	m := n
	if m > 8 {
		m = 8
	}
	var v uint64
	for i := uintptr(0); i < m; i++ {
		v |= uint64(*(*uint8)(unsafe.Add(base, first+i))) << (8 * i)
	}
	v >>= shift
	if n > 8 { // only possible when shift > 0
		v |= uint64(*(*uint8)(unsafe.Add(base, first+8))) << (64 - shift)
	}
	if width < 64 {
		v &= uint64(1)<<width - 1
	}
	return v
}

// Signed returns the same bits as Unsigned, sign-extended
// from bit width-1.
func Signed(base unsafe.Pointer, off, width uintptr) int64 {
	v := Unsigned(base, off, width)
	if width >= 64 {
		return int64(v)
	}
	s := 64 - width
	return int64(v<<s) >> s
}

// Set replaces bits off .. off+width-1 of the memory at base with
// the low width bits of v. All other bits are left unchanged.
func Set(base unsafe.Pointer, off, width uintptr, v uint64) {
	first := off >> 3
	n := ((off + width - 1) >> 3) - first + 1 // bytes touched, 1..9
	shift := off & 7

	mask := ^uint64(0)
	if width < 64 {
		mask = uint64(1)<<width - 1
	}
	v &= mask

	m := n
	if m > 8 {
		m = 8
	}
	var cur uint64
	for i := uintptr(0); i < m; i++ {
		cur |= uint64(*(*uint8)(unsafe.Add(base, first+i))) << (8 * i)
	}
	// Bits shifted out of the 64-bit window belong to the 9th byte.
	cur = cur&^(mask<<shift) | v<<shift
	for i := uintptr(0); i < m; i++ {
		*(*uint8)(unsafe.Add(base, first+i)) = uint8(cur >> (8 * i))
	}
	if n > 8 { // only possible when shift > 0
		hi := width - (64 - shift) // bits that land in the 9th byte, 1..7
		hmask := uint8(1)<<hi - 1
		p := (*uint8)(unsafe.Add(base, first+8))
		*p = *p&^hmask | uint8(v>>(64-shift))&hmask
	}
}
