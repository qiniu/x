package errors

import (
	"fmt"
	"os"
	"testing"
)

const (
	s1Exp = `not found

===> errors stack:
github.com/qiniu/x/errors.TestFrame()
	/Users/xushiwei/gop/x/errors/errors_test.go:11 errNotFound := New("not found")
github.com/qiniu/x/errors.TestFrame()
	/Users/xushiwei/gop/x/errors/errors_test.go:12 err1 := Frame(errNotFound, ...)
`

	s2Exp = `not found`
	s3Exp = `"not found"`
)

func TestFrame(t *testing.T) {
	file, _ := os.Getwd()
	file += "/errors_test.go"
	errNotFound := New("not found")
	err1 := NewFrame(errNotFound, `errNotFound := New("not found")`, file, 11, "github.com/qiniu/x/errors", "TestFrame", t)
	err2 := NewFrame(err1, `err1 := Frame(errNotFound, ...)`, file, 12, "github.com/qiniu/x/errors", "TestFrame", t)
	s1 := fmt.Sprint(err2)
	fmt.Println(s1)
	s2 := fmt.Sprintf("%s", err2)
	s3 := fmt.Sprintf("%q", err2)
	if s1 != s1Exp || s2 != s2Exp || s3 != s3Exp {
		t.Fatal("TestFrame failed:", s1, s2, s3)
	}
	if Unwrap(err2) != err1 {
		t.Fatal("Unwrap(err2) != err1")
	}
	if !Is(err2, errNotFound) {
		t.Fatal("!Is(err2, errNotFound)")
	}
	var err error
	_ = As(err, &err)
	_ = err2.NestedObject()
	_ = err2.ErrorDetail()
	_ = err2.AppendErrorDetail(nil)
	_ = err2.SummaryErr()
	_ = Detail(err2)
	_ = Info(err)
	_ = InfoEx(1, err)
	_ = err2.Detail(errNotFound)
}
