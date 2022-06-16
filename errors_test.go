package x_test

import (
	"strings"
	"syscall"
	"testing"

	"github.com/qiniu/x/errors"
)

// --------------------------------------------------------------------

func Foo() error {
	return syscall.ENOENT
}

func TestNewWith(t *testing.T) {
	err := Foo()
	if err != nil {
		err = errors.NewWith(err, `Foo()`, 2, "x_test.Foo")
		if strings.Index(err.Error(), "===> errors stack:\nx_test.Foo()\n\t") < 0 {
			t.Fatal(err)
		}
	}
}

// --------------------------------------------------------------------
