package objcache

import (
	"sync"
	"testing"
	"unsafe"
)

var (
	once        sync.Once
	stringGroup Getter

	// cacheFills is the number of times stringGroup or
	// protoGroup's Getter have been called. Read using the
	// cacheFills function.
	cacheFills AtomicInt
)

const (
	stringGroupName = "string-group"
)

type stringVal string

func (p stringVal) Dispose() error {
	return nil
}

func testSetup() {
	stringGroup = NewGroup(stringGroupName, 0, GetterFunc(func(key string) (val Value, err error) {
		cacheFills.Add(1)
		return stringVal("ECHO:" + key), nil
	}))
}

func countFills(f func()) int64 {
	fills0 := cacheFills.Get()
	f()
	return cacheFills.Get() - fills0
}

func TestCaching(t *testing.T) {
	once.Do(testSetup)
	fills := countFills(func() {
		for i := 0; i < 10; i++ {
			_, err := stringGroup.Get("TestCaching-key")
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	if fills != 1 {
		t.Errorf("expected 1 cache fill; got %d", fills)
	}
}

func TestGetVal(t *testing.T) {
	val, err := stringGroup.Get("short")
	if err != nil {
		t.Fatal(err)
	}
	if want := "ECHO:short"; string(val.(stringVal)) != want {
		t.Errorf("key got %q; want %q", val, want)
	}
}

func TestGroupStatsAlignment(t *testing.T) {
	var g Group
	off := unsafe.Offsetof(g.Stats)
	if off%8 != 0 {
		t.Fatal("Stats structure is not 8-byte aligned.")
	}
}
