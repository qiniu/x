package errors

import (
	"fmt"
	"testing"
)

const (
	s1Exp = `not found

===> errors stack:
github.com/qiniu/x/v8/errors.TestFrame("abcdefghijklmnop...", nil, 123, Array, Struct)
	/Users/xsw/qiniu/x/errors/errors_test.go:11 errNotFound := New("not found")
github.com/qiniu/x/v8/errors.TestFrame()
	/Users/xsw/qiniu/x/errors/errors_test.go:12 err1 := Frame(errNotFound, ...)
`

	s2Exp = `not found`
	s3Exp = `"not found"`
)

func TestFrame(t *testing.T) {
	file := "/Users/xsw/qiniu/x/errors/errors_test.go"
	errNotFound := New("not found")
	arg1 := "abcdefghijklmnopqrstuvwxyz"
	arg2 := interface{}(nil)
	arg3 := 123
	arg4 := [...]int{1}
	arg5 := struct{}{}
	err1 := NewFrame(errNotFound, `errNotFound := New("not found")`, file, 11, "github.com/qiniu/x/v8/errors", "TestFrame", arg1, arg2, arg3, arg4, arg5)
	err2 := NewFrame(err1, `err1 := Frame(errNotFound, ...)`, file, 12, "github.com/qiniu/x/v8/errors", "TestFrame")
	s1 := fmt.Sprint(err2)
	fmt.Println(s1)
	s2 := fmt.Sprintf("%s", err2)
	s3 := fmt.Sprintf("%q", err2)
	if s1 != s1Exp || s2 != s2Exp || s3 != s3Exp {
		t.Fatal("TestFrame failed:", s1, s2, s3)
	}
	_ = err2.NestedObject()
	_ = err2.ErrorDetail()
	_ = err2.AppendErrorDetail(nil)
	_ = err2.SummaryErr()
	_ = Detail(err2)
	_ = Info(err2)
	_ = InfoEx(1, err2)
	_ = err2.Detail(errNotFound)
}
