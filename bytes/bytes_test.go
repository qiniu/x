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

package bytes

import (
	"io"
	"testing"
)

// ---------------------------------------------------

func TestWriter(t *testing.T) {
	b := make([]byte, 8)
	w := NewWriter(b)
	w.Write([]byte("abc"))
	w.Write([]byte("1234567"))
	if _, e := w.Write([]byte("A")); e != io.EOF {
		t.Fatal("w.Write:", e)
	}
	if v := w.Len(); v != 8 {
		t.Fatal("w.Len():", v)
	}
	if v := string(w.Bytes()); v != "abc12345" {
		t.Fatal("w.Bytes():", v)
	}
	w.Reset()
	if v := w.Len(); v != 0 {
		t.Fatal("w.Len():", v)
	}
}

func TestBuffer(t *testing.T) {
	b := NewBuffer()
	n, err := b.WriteStringAt("Hello", 4)
	if n != 5 || err != nil {
		t.Fatal("WriteStringAt failed:", n, err)
	}
	if b.Len() != 9 {
		t.Fatal("Buffer.Len invalid (9 is required):", b.Len())
	}

	buf := make([]byte, 10)
	n, err = b.ReadAt(buf, 50)
	if n != 0 || err != io.EOF {
		t.Fatal("ReadAt failed:", n, err)
	}

	n, err = b.ReadAt(buf, 6)
	if n != 3 || err != io.EOF || string(buf[:n]) != "llo" {
		t.Fatal("ReadAt failed:", n, err, string(buf[:n]))
	}

	n, err = b.WriteAt([]byte("Hi h"), 1)
	if n != 4 || err != nil {
		t.Fatal("WriteAt failed:", n, err)
	}
	if b.Len() != 9 {
		t.Fatal("Buffer.Len invalid (9 is required):", b.Len())
	}

	n, err = b.ReadAt(buf, 0)
	if n != 9 || err != io.EOF || string(buf[:n]) != "\x00Hi hello" {
		t.Fatal("ReadAt failed:", n, err)
	}

	n, err = b.WriteStringAt("LO world!", 7)
	if n != 9 || err != nil {
		t.Fatal("WriteStringAt failed:", n, err)
	}
	if b.Len() != 16 {
		t.Fatal("Buffer.Len invalid (16 is required):", b.Len())
	}

	buf = make([]byte, 17)
	n, err = b.ReadAt(buf, 0)
	if n != 16 || err != io.EOF || string(buf[:n]) != "\x00Hi helLO world!" {
		t.Fatal("ReadAt failed:", n, err, string(buf[:n]))
	}

	off := int64(b.Len())
	b.Truncate(off + 1)
	if n, err = b.ReadAt(buf[:1], off); n != 1 || err != nil || buf[0] != 0 {
		t.Fatal("b.Truncate(off+1) failed:", n, err, buf[0])
	}
	if err := b.Truncate(1); err != nil || len(b.Buffer()) != 1 {
		t.Fatal("b.Truncate(1) failed:", err)
	}
	b.WriteStringAt("123", 1)
	if string(b.Buffer()) != "\x00123" {
		t.Fatal("b.WriteStringAt 123:", string(b.Buffer()))
	}
	b.WriteAt([]byte("Hello!!! "), 0)
	if string(b.Buffer()) != "Hello!!! " {
		t.Fatal("b.WriteAt Hello!!! :", string(b.Buffer()))
	}
	b.WriteAt([]byte(" "), 9)
	if string(b.Buffer()) != "Hello!!!  " {
		t.Fatal("b.WriteAt :", string(b.Buffer()))
	}
}

// ---------------------------------------------------
