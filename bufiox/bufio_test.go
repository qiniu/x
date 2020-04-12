package bufiox

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
	"unsafe"
)

// -------------------------------------------------------------------------------------

func TestReaderSize(t *testing.T) {
	b1 := NewReaderBuffer(nil)
	size1 := unsafe.Sizeof(*b1)
	b2 := bufio.NewReader(os.Stdin)
	size2 := unsafe.Sizeof(*b2)
	if size1 != size2 {
		t.Fatal("TestReaderSize: sizeof(bufiox.Reader) != sizeof(bufio.Reader)")
	}
	b, err := ReadAll(b1)
	if err != nil || b != nil {
		t.Fatal("ReadAll failed:", err, b)
	}
}

func TestGetUnderlyingReaderAndSeek(t *testing.T) {
	r := strings.NewReader("Hello, china!!!")
	b := bufio.NewReader(r)
	if getUnderlyingReader(b) != r {
		t.Fatal("getUnderlyingReader failed")
	}
	b.ReadByte()
	r1 := getUnderlyingReader(b)
	b1 := bufio.NewReader(r1)
	if _, err1 := b1.ReadByte(); err1 != io.EOF {
		t.Fatal("bufio.NewReader cache?")
	}
	newoff, err := Seek(b, 7, io.SeekStart)
	if err != nil || newoff != 7 {
		t.Fatal("Seek failed:", err, newoff)
	}
	china, err := b.ReadString('!')
	if err != nil {
		t.Fatal("ReadString failed:", err)
	}
	if china != "china!" {
		t.Fatal("Seek failed:", china)
	}
}

// -------------------------------------------------------------------------------------

func TestSeeker(t *testing.T) {
	r := strings.NewReader("Hello, china!!!")
	b := NewReader(r)
	var rdseeker io.ReadSeeker = b
	rdseeker.Seek(7, io.SeekStart)
	china, err := b.ReadString('!')
	if err != nil {
		t.Fatal("ReadString failed:", err)
	}
	if china != "china!" {
		t.Fatal("Seek failed:", china)
	}
	b.Seek(0, io.SeekStart)

	b2 := NewReaderSize(r, 64)
	data, err := ReadAll(&b2.Reader)
	if err != nil || string(data) != "Hello, china!!!" {
		t.Fatal("ReadAll failed:", err, data)
	}

	b3 := NewReaderSize(b2, 32)
	if b2 != b3 {
		t.Fatal("NewReader on *bufiox.Reader")
	}
}

// -------------------------------------------------------------------------------------
