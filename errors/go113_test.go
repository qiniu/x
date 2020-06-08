// +build go1.13

package errors

import (
	"testing"
)

func TestGo113(t *testing.T) {
	file := "/Users/xsw/qiniu/x/errors/errors_test.go"
	errNotFound := New("not found")
	err1 := NewFrame(errNotFound, `errNotFound := New("not found")`, file, 11, "github.com/qiniu/x/v8/errors", "TestFrame", t)
	err2 := NewFrame(err1, `err1 := Frame(errNotFound, ...)`, file, 12, "github.com/qiniu/x/v8/errors", "TestFrame", t)
	if Unwrap(err2) != err1 {
		t.Fatal("Unwrap(err2) != err1")
	}
	if !Is(err2, errNotFound) {
		t.Fatal("!Is(err2, errNotFound)")
	}
	var err error
	_ = As(err, &err)
}
