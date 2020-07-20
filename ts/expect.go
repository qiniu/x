package ts

import (
	"bytes"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/qiniu/x/errors"
)

// ----------------------------------------------------------------------------

type captureData struct {
	target **os.File
	old    *os.File
	exp    chan interface{}
	done   chan bool
}

func (p *captureData) expect(t *testing.T, v interface{}) {
	p.exp <- v
	if equal := <-p.done; !equal {
		t.FailNow()
	}
}

// Expecting represents a expecting object.
type Expecting struct {
	t     *testing.T
	cdata [2]*captureData
	msg   []byte
	rcov  interface{}
	cstk  *stack
}

const (
	// CapStdout - capture stdout
	CapStdout = 1
	// CapStderr - capture stderr
	CapStderr = 2
	// CapOutput - capture stdout and stderr
	CapOutput = CapStdout | CapStderr
)

// StartExpecting starts expecting with a capture mode.
func StartExpecting(t *testing.T, mode int) *Expecting {
	p := &Expecting{t: t}
	if mode == 0 {
		mode = CapOutput
	}
	if (mode & CapStdout) != 0 {
		p.capture(0, &os.Stdout)
	}
	if (mode & CapStderr) != 0 {
		p.capture(1, &os.Stderr)
	}
	return p
}

// Call creates a test case, and then calls a function.
func (p *Expecting) Call(fn interface{}, args ...interface{}) (e *Expecting) {
	e = p
	e.msg = errors.CallDetail(nil, fn, args...)
	defer func() {
		if e.rcov = recover(); e.rcov != nil {
			e.cstk = callers(3)
		}
	}()
	e.rcov = nil
	reflect.ValueOf(fn).Call(makeArgs(args))
	return
}

// Expect expects stdout ouput is equal to text.
func (p *Expecting) Expect(text interface{}) *Expecting {
	p.cdata[0].expect(p.t, text)
	return p
}

// ExpectErr expects stderr ouput is equal to text.
func (p *Expecting) ExpectErr(text interface{}) *Expecting {
	p.cdata[1].expect(p.t, text)
	return p
}

// NoPanic indicates that no panic occurs.
func (p *Expecting) NoPanic() {
	if p.rcov != nil {
		p.t.Fatalf("panic: %v\n%+v\n", p.rcov, p.cstk)
	}
}

// Panic checks if function call panics or not. Panic(v) means
// function call panics with `v`. If v == nil, it means we don't
// care any detail information about panic. Panic() indicates
// that no panic occurs.
func (p *Expecting) Panic(panicMsg ...interface{}) *Expecting {
	if panicMsg == nil {
		p.NoPanic()
	} else {
		assertPanic(p.t, p.msg, p.rcov, panicMsg[0])
	}
	return p
}

// Close stops expecting.
func (p *Expecting) Close() error {
	for i, cdata := range p.cdata {
		if cdata != nil {
			p.cdata[i] = nil
			w := *cdata.target
			*cdata.target = cdata.old
			w.Close()
			cdata.exp <- nil // close pipe
		}
	}
	return nil
}

func (p *Expecting) capture(idx int, target **os.File) {
	old := *target
	cdata := &captureData{
		target: target,
		old:    old,
		exp:    make(chan interface{}),
		done:   make(chan bool, 1),
	}
	p.cdata[idx] = cdata
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	var mutex sync.Mutex
	var buf = bytes.NewBuffer(nil)
	go func() {
		b := make([]byte, 1024)
		for {
			n, err := r.Read(b)
			if err != nil {
				return
			}
			if n > 0 {
				mutex.Lock()
				buf.Write(b[:n])
				old.Write(b[:n])
				mutex.Unlock()
			}
		}
	}()
	go func() {
		for exp := range cdata.exp {
			if exp == nil { // close pipe
				r.Close()
				return
			}
			deadline := time.Now().Add(time.Second / 10) // 0.1s
			off, b := 0, toString(exp)
		retry:
			mutex.Lock()
			a := string(buf.Bytes())
			buf.Reset()
			mutex.Unlock()
			equal := (a == b[off:])
			if !equal {
				if strings.HasPrefix(b[off:], a) && time.Now().Before(deadline) {
					off += len(a)
					runtime.Gosched()
					goto retry
				}
				*target = old
				p.t.Logf("%s:\n%s\nExpect:\n%s\n", string(p.msg), b[:off]+a, b)
				*target = w
			}
			cdata.done <- equal
		}
	}()
	*target = w
}

func toString(exp interface{}) string {
	switch v := exp.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	default:
		panic("expect: unsupport type - " + reflect.TypeOf(exp).String())
	}
}

// ----------------------------------------------------------------------------
