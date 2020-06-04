// +build go1.13

package errors

import (
	"os"
	"testing"
)

func TestGo113(t *testing.T) {
	file, _ := os.Getwd()
	file += "/errors_test.go"
	errNotFound := New("not found")
	err1 := NewFrame(errNotFound, `errNotFound := New("not found")`, file, 11, "github.com/qiniu/x/errors", "TestFrame", t)
	err2 := NewFrame(err1, `err1 := Frame(errNotFound, ...)`, file, 12, "github.com/qiniu/x/errors", "TestFrame", t)
	if Unwrap(err2) != err1 {
		t.Fatal("Unwrap(err2) != err1")
	}
	if !Is(err2, errNotFound) {
		t.Fatal("!Is(err2, errNotFound)")
	}
	var err error
	_ = As(err, &err)
}
