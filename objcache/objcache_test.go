package objcache

import (
	"sync"
	"sync/atomic"
	"testing"
)

var (
	once        sync.Once
	stringGroup *Group
	cacheFills  int64
)

const (
	stringGroupName = "string-group"
)

type stringVal string

func (p stringVal) Dispose() error {
	return nil
}

func testSetup() {
	stringGroup = NewGroup(stringGroupName, 0, func(ctx Context, key Key) (val Value, err error) {
		atomic.AddInt64(&cacheFills, 1)
		return stringVal("ECHO:" + key.(string)), nil
	})
}

func countFills(f func()) int64 {
	fills0 := atomic.LoadInt64(&cacheFills)
	f()
	return atomic.LoadInt64(&cacheFills) - fills0
}

func TestCaching(t *testing.T) {
	once.Do(testSetup)
	fills := countFills(func() {
		for i := 0; i < 10; i++ {
			_, err := stringGroup.Get(nil, "TestCaching-key")
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
	val, err := stringGroup.Get(nil, "short")
	if err != nil {
		t.Fatal(err)
	}
	if want := "ECHO:short"; string(val.(stringVal)) != want {
		t.Errorf("key got %q; want %q", val, want)
	}
}
