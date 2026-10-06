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

package bitfield

import (
	"math/rand"
	"testing"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Reference implementation: one bit at a time, straight from the definition
// "bit k is bit (k % 8) of byte (k / 8)".

func refUnsigned(b []byte, off, w int) uint64 {
	var v uint64
	for i := 0; i < w; i++ {
		k := off + i
		if b[k/8]>>(k%8)&1 == 1 {
			v |= 1 << i
		}
	}
	return v
}

func refSigned(b []byte, off, w int) int64 {
	v := refUnsigned(b, off, w)
	if w < 64 && v>>(w-1)&1 == 1 {
		v |= ^uint64(0) << w
	}
	return int64(v)
}

func refSet(b []byte, off, w int, v uint64) {
	for i := 0; i < w; i++ {
		k := off + i
		if v>>i&1 == 1 {
			b[k/8] |= 1 << (k % 8)
		} else {
			b[k/8] &^= 1 << (k % 8)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers

// guarded is a buffer whose usable area is surrounded by guard bytes, so that
// tests can detect any access outside the area.
type guarded struct {
	buf  []byte
	size int
}

const guardLen = 16

func newGuarded(size int, fill byte) *guarded {
	g := &guarded{buf: make([]byte, size+2*guardLen), size: size}
	for i := range g.buf {
		g.buf[i] = fill
	}
	return g
}

func (g *guarded) base() unsafe.Pointer { return unsafe.Pointer(&g.buf[guardLen]) }
func (g *guarded) area() []byte         { return g.buf[guardLen : guardLen+g.size] }

func (g *guarded) checkGuards(t *testing.T, fill byte) {
	t.Helper()
	for i := 0; i < guardLen; i++ {
		if g.buf[i] != fill || g.buf[guardLen+g.size+i] != fill {
			t.Fatalf("guard byte modified (front %d / back %d)", i, i)
		}
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Fixed cases

func TestUnsignedFixed(t *testing.T) {
	tests := []struct {
		name  string
		mem   []byte
		off   uintptr
		width uintptr
		want  uint64
	}{
		{"1 bit, bit 0", []byte{0x01}, 0, 1, 1},
		{"1 bit, bit 7", []byte{0x80}, 7, 1, 1},
		{"1 bit, unset", []byte{0xfe}, 0, 1, 0},
		{"low nibble", []byte{0xab}, 0, 4, 0xb},
		{"high nibble", []byte{0xab}, 4, 4, 0xa},
		{"full byte", []byte{0xab}, 0, 8, 0xab},
		{"second byte", []byte{0x00, 0xcd}, 8, 8, 0xcd},
		{"straddle two bytes", []byte{0xf0, 0x0f}, 4, 8, 0xff},
		{"straddle, mixed", []byte{0xa0, 0x05}, 4, 8, 0x5a},
		{"16 bits little-endian", []byte{0x34, 0x12}, 0, 16, 0x1234},
		{"32 bits", []byte{0x78, 0x56, 0x34, 0x12}, 0, 32, 0x12345678},
		{"3 bits at offset 5", []byte{0xe0}, 5, 3, 7},
		{"width 63 aligned", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, 63, 1<<63 - 1},
		{"width 64 aligned", []byte{1, 2, 3, 4, 5, 6, 7, 8}, 0, 64, 0x0807060504030201},
		{"ignores bits outside range", []byte{0xff, 0xff}, 3, 2, 3},
		{"zero bits stay zero", []byte{0x00, 0x00, 0x00}, 5, 11, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unsigned(unsafe.Pointer(&tt.mem[0]), tt.off, tt.width); got != tt.want {
				t.Fatalf("Unsigned(off=%d, width=%d) = %#x, want %#x", tt.off, tt.width, got, tt.want)
			}
		})
	}
}

func TestSignedFixed(t *testing.T) {
	tests := []struct {
		name  string
		mem   []byte
		off   uintptr
		width uintptr
		want  int64
	}{
		{"1 bit set is -1", []byte{0x01}, 0, 1, -1},
		{"1 bit clear is 0", []byte{0x00}, 0, 1, 0},
		{"4 bits 0b1101 is -3", []byte{0xd0}, 4, 4, -3},
		{"4 bits 0b0111 is 7", []byte{0x07}, 0, 4, 7},
		{"4 bits 0b1000 is -8", []byte{0x08}, 0, 4, -8},
		{"8 bits 0xff is -1", []byte{0xff}, 0, 8, -1},
		{"8 bits 0x80 is -128", []byte{0x80}, 0, 8, -128},
		{"8 bits 0x7f is 127", []byte{0x7f}, 0, 8, 127},
		{"12 bits straddling", []byte{0xf0, 0xff}, 4, 12, -1},
		{"16 bits", []byte{0x00, 0x80}, 0, 16, -32768},
		{"32 bits -1", []byte{0xff, 0xff, 0xff, 0xff}, 0, 32, -1},
		{"63 bits all ones", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, 63, -1},
		{"64 bits min int", []byte{0, 0, 0, 0, 0, 0, 0, 0x80}, 0, 64, -1 << 63},
		{"64 bits -1", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, 64, -1},
		{"sign bit only decides, rest ignored", []byte{0xff, 0x00}, 6, 4, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Signed(unsafe.Pointer(&tt.mem[0]), tt.off, tt.width); got != tt.want {
				t.Fatalf("Signed(off=%d, width=%d) = %d, want %d", tt.off, tt.width, got, tt.want)
			}
		})
	}
}

func TestSetFixed(t *testing.T) {
	tests := []struct {
		name  string
		mem   []byte
		off   uintptr
		width uintptr
		v     uint64
		want  []byte
	}{
		{"set single bit", []byte{0x00}, 3, 1, 1, []byte{0x08}},
		{"clear single bit", []byte{0xff}, 3, 1, 0, []byte{0xf7}},
		{"set low nibble keeps high", []byte{0xf0}, 0, 4, 0x5, []byte{0xf5}},
		{"set high nibble keeps low", []byte{0x0f}, 4, 4, 0xa, []byte{0xaf}},
		{"overwrite with zeros", []byte{0xff, 0xff}, 4, 8, 0, []byte{0x0f, 0xf0}},
		{"overwrite with ones", []byte{0x00, 0x00}, 4, 8, 0xff, []byte{0xf0, 0x0f}},
		{"full byte", []byte{0x12, 0x34, 0x56}, 8, 8, 0xab, []byte{0x12, 0xab, 0x56}},
		{"16 bits aligned", []byte{0xff, 0xff, 0xff, 0xff}, 8, 16, 0x1234, []byte{0xff, 0x34, 0x12, 0xff}},
		{"64 bits aligned", make([]byte, 8), 0, 64, 0x0807060504030201, []byte{1, 2, 3, 4, 5, 6, 7, 8}},
		{"value truncated to width", []byte{0x00}, 0, 3, 0xff, []byte{0x07}},
		{"truncation keeps neighbours", []byte{0x00}, 2, 3, 0xff, []byte{0x1c}},
		{"negative value two's complement", []byte{0x00}, 4, 4, uint64(0xfffffffffffffffd), []byte{0xd0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Set(unsafe.Pointer(&tt.mem[0]), tt.off, tt.width, tt.v)
			if !bytesEqual(tt.mem, tt.want) {
				t.Fatalf("Set(off=%d, width=%d, %#x): got % x, want % x", tt.off, tt.width, tt.v, tt.mem, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Fields of up to 64 bits that are not byte-aligned span 9 bytes.

func TestNineByteSpan(t *testing.T) {
	for shift := 1; shift <= 7; shift++ {
		w := 64
		off := shift
		t.Run("shift"+string(rune('0'+shift)), func(t *testing.T) {
			g := newGuarded(9, 0x00)
			const want = 0xfedcba9876543210
			Set(g.base(), uintptr(off), uintptr(w), want)
			if got := Unsigned(g.base(), uintptr(off), uintptr(w)); got != want {
				t.Fatalf("round trip: got %#x, want %#x", got, uint64(want))
			}
			ref := make([]byte, 9)
			refSet(ref, off, w, want)
			if !bytesEqual(g.area(), ref) {
				t.Fatalf("memory: got % x, want % x", g.area(), ref)
			}
			// Bits below off and above off+w must stay 0.
			if g.area()[0]&(1<<shift-1) != 0 || g.area()[8]>>shift != 0 {
				t.Fatalf("bits outside the field were touched: % x", g.area())
			}
			g.checkGuards(t, 0x00)
		})
	}
}

func TestNineByteSpanPreservesNeighbours(t *testing.T) {
	for shift := 1; shift <= 7; shift++ {
		g := newGuarded(9, 0xff)
		Set(g.base(), uintptr(shift), 64, 0)
		ref := bytesOf(0xff, 9)
		refSet(ref, shift, 64, 0)
		if !bytesEqual(g.area(), ref) {
			t.Fatalf("shift %d: got % x, want % x", shift, g.area(), ref)
		}
		if got := Unsigned(g.base(), uintptr(shift), 64); got != 0 {
			t.Fatalf("shift %d: Unsigned = %#x, want 0", shift, got)
		}
		g.checkGuards(t, 0xff)
	}
}

func TestSignedNineByteSpan(t *testing.T) {
	for shift := 0; shift <= 7; shift++ {
		g := newGuarded(9, 0x55)
		Set(g.base(), uintptr(shift), 64, uint64(1<<63)) // min int64
		if got := Signed(g.base(), uintptr(shift), 64); got != -1<<63 {
			t.Fatalf("shift %d: got %d", shift, got)
		}
		Set(g.base(), uintptr(shift), 64, ^uint64(0))
		if got := Signed(g.base(), uintptr(shift), 64); got != -1 {
			t.Fatalf("shift %d: got %d", shift, got)
		}
		g.checkGuards(t, 0x55)
	}
}

func bytesOf(b byte, n int) []byte {
	s := make([]byte, n)
	for i := range s {
		s[i] = b
	}
	return s
}

// ---------------------------------------------------------------------------
// Semantics

// Set followed by Unsigned returns the value truncated to width bits.
func TestSetUnsignedRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for it := 0; it < 20000; it++ {
		w := 1 + r.Intn(64)
		off := r.Intn(8*16 - w + 1)
		v := r.Uint64()
		g := newGuarded(16, byte(r.Intn(256)))
		fill := g.buf[0]
		Set(g.base(), uintptr(off), uintptr(w), v)
		want := v
		if w < 64 {
			want &= 1<<w - 1
		}
		if got := Unsigned(g.base(), uintptr(off), uintptr(w)); got != want {
			t.Fatalf("off=%d w=%d: got %#x, want %#x", off, w, got, want)
		}
		g.checkGuards(t, fill)
	}
}

// Writing a signed value and reading it back sign-extended gives the value
// again, as long as it fits in width bits.
func TestSignedRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for it := 0; it < 20000; it++ {
		w := 1 + r.Intn(64)
		off := r.Intn(8*16 - w + 1)
		var v int64
		if w == 64 {
			v = int64(r.Uint64())
		} else {
			lo := -(int64(1) << (w - 1))
			span := uint64(1) << w
			v = lo + int64(r.Uint64()%span)
		}
		g := newGuarded(16, 0xa5)
		Set(g.base(), uintptr(off), uintptr(w), uint64(v))
		if got := Signed(g.base(), uintptr(off), uintptr(w)); got != v {
			t.Fatalf("off=%d w=%d: got %d, want %d", off, w, got, v)
		}
		g.checkGuards(t, 0xa5)
	}
}

// Setting the same value twice leaves memory unchanged the second time.
func TestSetIdempotent(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for it := 0; it < 5000; it++ {
		w := 1 + r.Intn(64)
		off := r.Intn(8*16 - w + 1)
		v := r.Uint64()
		g := newGuarded(16, 0x3c)
		Set(g.base(), uintptr(off), uintptr(w), v)
		first := append([]byte(nil), g.area()...)
		Set(g.base(), uintptr(off), uintptr(w), v)
		if !bytesEqual(first, g.area()) {
			t.Fatalf("off=%d w=%d: second Set changed memory", off, w)
		}
	}
}

// Fields that sit next to each other do not disturb one another.
func TestAdjacentFieldsIndependent(t *testing.T) {
	widths := []uintptr{1, 3, 4, 7, 8, 9, 13, 16, 5, 64, 2, 11}
	var offs []uintptr
	var total uintptr
	for _, w := range widths {
		offs = append(offs, total)
		total += w
	}
	g := newGuarded(int((total+7)/8), 0x00)
	r := rand.New(rand.NewSource(5))
	want := make([]uint64, len(widths))
	for round := 0; round < 200; round++ {
		i := r.Intn(len(widths))
		v := r.Uint64()
		if widths[i] < 64 {
			v &= 1<<widths[i] - 1
		}
		want[i] = v
		Set(g.base(), offs[i], widths[i], v)
		for j := range widths {
			if got := Unsigned(g.base(), offs[j], widths[j]); got != want[j] {
				t.Fatalf("round %d: field %d (off=%d w=%d) = %#x, want %#x",
					round, j, offs[j], widths[j], got, want[j])
			}
		}
	}
	g.checkGuards(t, 0x00)
}

// Only the bytes that hold bits of the range may be read or written.
func TestOnlyTouchesNeededBytes(t *testing.T) {
	for w := 1; w <= 64; w++ {
		for off := 0; off < 24; off++ {
			n := (off+w-1)/8 + 1 // bytes covering the range
			g := newGuarded(n, 0xaa)
			Set(g.base(), uintptr(off), uintptr(w), ^uint64(0))
			_ = Unsigned(g.base(), uintptr(off), uintptr(w))
			_ = Signed(g.base(), uintptr(off), uintptr(w))
			g.checkGuards(t, 0xaa) // would fail on any access beyond [0, n)
			if got := Unsigned(g.base(), uintptr(off), uintptr(w)); got != maskOf(w) {
				t.Fatalf("off=%d w=%d: got %#x, want %#x", off, w, got, maskOf(w))
			}
		}
	}
}

func maskOf(w int) uint64 {
	if w >= 64 {
		return ^uint64(0)
	}
	return 1<<w - 1
}

// ---------------------------------------------------------------------------
// Exhaustive and randomized comparison with the reference implementation.

func TestExhaustiveSmall(t *testing.T) {
	// Every (off, width) in a 4-byte window, over a few memory patterns.
	patterns := [][]byte{
		{0x00, 0x00, 0x00, 0x00},
		{0xff, 0xff, 0xff, 0xff},
		{0xaa, 0x55, 0xaa, 0x55},
		{0x01, 0x02, 0x04, 0x08},
		{0xde, 0xad, 0xbe, 0xef},
	}
	values := []uint64{0, 1, 2, 0x55555555, 0xaaaaaaaa, 0xffffffff, ^uint64(0)}
	for _, pat := range patterns {
		for w := 1; w <= 32; w++ {
			for off := 0; off+w <= 32; off++ {
				mem := append([]byte(nil), pat...)
				base := unsafe.Pointer(&mem[0])
				if got, want := Unsigned(base, uintptr(off), uintptr(w)), refUnsigned(mem, off, w); got != want {
					t.Fatalf("Unsigned pat=% x off=%d w=%d: got %#x want %#x", pat, off, w, got, want)
				}
				if got, want := Signed(base, uintptr(off), uintptr(w)), refSigned(mem, off, w); got != want {
					t.Fatalf("Signed pat=% x off=%d w=%d: got %d want %d", pat, off, w, got, want)
				}
				for _, v := range values {
					m1 := append([]byte(nil), pat...)
					m2 := append([]byte(nil), pat...)
					Set(unsafe.Pointer(&m1[0]), uintptr(off), uintptr(w), v)
					refSet(m2, off, w, v)
					if !bytesEqual(m1, m2) {
						t.Fatalf("Set pat=% x off=%d w=%d v=%#x: got % x want % x", pat, off, w, v, m1, m2)
					}
				}
			}
		}
	}
}

func TestAllWidthsAllOffsets(t *testing.T) {
	// Every width 1..64 at every offset within a 24-byte window,
	// including all 9-byte spans.
	r := rand.New(rand.NewSource(6))
	for w := 1; w <= 64; w++ {
		for off := 0; off+w <= 8*24; off++ {
			g := newGuarded(24, 0x00)
			r.Read(g.area())
			ref := append([]byte(nil), g.area()...)
			if got, want := Unsigned(g.base(), uintptr(off), uintptr(w)), refUnsigned(ref, off, w); got != want {
				t.Fatalf("Unsigned off=%d w=%d: got %#x want %#x", off, w, got, want)
			}
			if got, want := Signed(g.base(), uintptr(off), uintptr(w)), refSigned(ref, off, w); got != want {
				t.Fatalf("Signed off=%d w=%d: got %d want %d", off, w, got, want)
			}
			v := r.Uint64()
			Set(g.base(), uintptr(off), uintptr(w), v)
			refSet(ref, off, w, v)
			if !bytesEqual(g.area(), ref) {
				t.Fatalf("Set off=%d w=%d v=%#x: got % x want % x", off, w, v, g.area(), ref)
			}
			g.checkGuards(t, 0x00)
		}
	}
}

func TestRandomAgainstReference(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for it := 0; it < 300000; it++ {
		g := newGuarded(16, 0x00)
		r.Read(g.area())
		ref := append([]byte(nil), g.area()...)
		w := 1 + r.Intn(64)
		off := r.Intn(8*16 - w + 1)
		if got, want := Unsigned(g.base(), uintptr(off), uintptr(w)), refUnsigned(ref, off, w); got != want {
			t.Fatalf("Unsigned off=%d w=%d: got %#x want %#x", off, w, got, want)
		}
		if got, want := Signed(g.base(), uintptr(off), uintptr(w)), refSigned(ref, off, w); got != want {
			t.Fatalf("Signed off=%d w=%d: got %d want %d", off, w, got, want)
		}
		v := r.Uint64()
		Set(g.base(), uintptr(off), uintptr(w), v)
		refSet(ref, off, w, v)
		if !bytesEqual(g.area(), ref) {
			t.Fatalf("Set off=%d w=%d v=%#x: got % x want % x", off, w, v, g.area(), ref)
		}
		g.checkGuards(t, 0x00)
	}
}

// ---------------------------------------------------------------------------
// The examples from the llcppg bit-field proposal, written the way generated
// code would use the helpers.

// Packet mirrors:
//
//	typedef struct {
//	    unsigned int ready : 1;
//	    unsigned int mode  : 3;
//	    int          delta : 4;
//	    unsigned int       : 0;
//	    unsigned int level : 6;
//	    int          id;
//	} Packet;
type Packet struct {
	_xgo_bits_0 [5]uint8
	_           [3]uint8 // padding so that Id sits at byte 8
	Id          int32
}

func (p *Packet) getReady() uint32  { return uint32(Unsigned(unsafe.Pointer(p), 0, 1)) }
func (p *Packet) setReady(v uint32) { Set(unsafe.Pointer(p), 0, 1, uint64(v)) }
func (p *Packet) getMode() uint32   { return uint32(Unsigned(unsafe.Pointer(p), 1, 3)) }
func (p *Packet) setMode(v uint32)  { Set(unsafe.Pointer(p), 1, 3, uint64(v)) }
func (p *Packet) getDelta() int32   { return int32(Signed(unsafe.Pointer(p), 4, 4)) }
func (p *Packet) setDelta(v int32)  { Set(unsafe.Pointer(p), 4, 4, uint64(v)) }
func (p *Packet) getLevel() uint32  { return uint32(Unsigned(unsafe.Pointer(p), 32, 6)) }
func (p *Packet) setLevel(v uint32) { Set(unsafe.Pointer(p), 32, 6, uint64(v)) }

func TestPacketLayout(t *testing.T) {
	var p Packet
	if got := unsafe.Sizeof(p); got != 12 {
		t.Fatalf("sizeof(Packet) = %d, want 12", got)
	}
	if got := unsafe.Offsetof(p.Id); got != 8 {
		t.Fatalf("offsetof(Id) = %d, want 8", got)
	}
}

func TestPacket(t *testing.T) {
	var p Packet
	p.setReady(1)
	p.setMode(5)
	p.setDelta(-3)
	p.setLevel(40)
	p.Id = 42

	if p.getReady() != 1 || p.getMode() != 5 || p.getDelta() != -3 || p.getLevel() != 40 || p.Id != 42 {
		t.Fatalf("got ready=%d mode=%d delta=%d level=%d id=%d",
			p.getReady(), p.getMode(), p.getDelta(), p.getLevel(), p.Id)
	}
	// ready=1, mode=5 (101), delta=-3 (1101): low byte is 1101 101 1 = 0xdb.
	if p._xgo_bits_0[0] != 0xdb {
		t.Fatalf("byte 0 = %#x, want 0xdb", p._xgo_bits_0[0])
	}
	// level=40 (101000) sits in byte 4; the rest of that byte stays 0.
	if p._xgo_bits_0[4] != 40 {
		t.Fatalf("byte 4 = %#x, want 0x28", p._xgo_bits_0[4])
	}
	// Bytes 1..3 belong to the skipped allocation unit and must stay 0.
	if p._xgo_bits_0[1] != 0 || p._xgo_bits_0[2] != 0 || p._xgo_bits_0[3] != 0 {
		t.Fatalf("unused bytes were modified: %v", p._xgo_bits_0)
	}
}

func TestPacketFieldsDoNotInterfere(t *testing.T) {
	var p Packet
	p.setMode(7)
	p.setReady(1)
	p.setReady(0)
	if p.getMode() != 7 {
		t.Fatalf("mode = %d after toggling ready, want 7", p.getMode())
	}
	p.setDelta(-8)
	p.setMode(0)
	if p.getDelta() != -8 {
		t.Fatalf("delta = %d after clearing mode, want -8", p.getDelta())
	}
	p.setLevel(63)
	p.setLevel(1)
	if p.getLevel() != 1 {
		t.Fatalf("level = %d, want 1", p.getLevel())
	}
	p.Id = -1
	p.setLevel(0)
	if p.Id != -1 {
		t.Fatalf("Id = %d, want -1", p.Id)
	}
}

func TestPacketTruncationAndNegative(t *testing.T) {
	var p Packet
	p.setMode(0xf) // 4 bits into a 3-bit field: stores 0b111
	if p.getMode() != 7 || p.getReady() != 0 || p.getDelta() != 0 {
		t.Fatalf("mode=%d ready=%d delta=%d", p.getMode(), p.getReady(), p.getDelta())
	}
	p.setDelta(-3)
	if p._xgo_bits_0[0]>>4 != 0b1101 {
		t.Fatalf("delta bits = %#b, want 0b1101", p._xgo_bits_0[0]>>4)
	}
	p.setDelta(9) // 0b1001 in 4 bits is -7
	if p.getDelta() != -7 {
		t.Fatalf("delta = %d, want -7", p.getDelta())
	}
}

// Perm mirrors a struct whose three 1-bit fields all live in one byte.
type Perm struct {
	_xgo_align  [0]uint32
	_xgo_bits_0 [1]uint8
}

func TestPerm(t *testing.T) {
	var p Perm
	if unsafe.Sizeof(p) != 4 || unsafe.Alignof(p) != 4 {
		t.Fatalf("size=%d align=%d, want 4 and 4", unsafe.Sizeof(p), unsafe.Alignof(p))
	}
	Set(unsafe.Pointer(&p), 0, 1, 1) // r
	Set(unsafe.Pointer(&p), 2, 1, 1) // x
	if p._xgo_bits_0[0] != 0b101 {
		t.Fatalf("bits = %#b, want 0b101", p._xgo_bits_0[0])
	}
	if Unsigned(unsafe.Pointer(&p), 1, 1) != 0 {
		t.Fatal("w should be 0")
	}
}

// A bool bit-field: the getter reports non-zero, the setter stores 0 or 1.
func TestBoolStyle(t *testing.T) {
	var b [1]byte
	base := unsafe.Pointer(&b[0])
	flag := func() bool { return Unsigned(base, 5, 1) != 0 }
	setFlag := func(v bool) {
		var x uint64
		if v {
			x = 1
		}
		Set(base, 5, 1, x)
	}
	if flag() {
		t.Fatal("flag should start false")
	}
	setFlag(true)
	if !flag() || b[0] != 0x20 {
		t.Fatalf("after true: flag=%v byte=%#x", flag(), b[0])
	}
	setFlag(false)
	if flag() || b[0] != 0 {
		t.Fatalf("after false: flag=%v byte=%#x", flag(), b[0])
	}
}

// A bit-field wider than 32 bits and not byte aligned, as C++ allows for
// 64-bit declared types.
func TestWideField(t *testing.T) {
	var m [16]byte
	base := unsafe.Pointer(&m[0])
	Set(base, 3, 40, 0xab_cdef_0123)
	if got := Unsigned(base, 3, 40); got != 0xab_cdef_0123 {
		t.Fatalf("got %#x", got)
	}
	if got := Signed(base, 3, 40); got != 0xab_cdef_0123-1<<40 {
		t.Fatalf("signed got %d", got)
	}
}
