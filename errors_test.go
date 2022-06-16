package x

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
		err = errors.NewWith(err, 2, `Foo()`, "x", "Foo")
		if strings.Index(err.Error(), "===> errors stack:\nx.Foo()\n\t") < 0 {
			t.Fatal(err)
		}
	}
}

// --------------------------------------------------------------------
