package errors

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

// --------------------------------------------------------------------

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(msg string) error {
	return errors.New(msg)
}

// Err returns the cause error.
func Err(err error) error {
	if e, ok := err.(*Frame); ok {
		return Err(e.Err)
	}
	return err
}

// --------------------------------------------------------------------

// Frame represents an error frame.
type Frame struct {
	Err  error
	Pkg  string
	Func string
	Args []interface{}
	Code string
	File string
	Line int
}

// NewFrame creates a new error frame.
func NewFrame(err error, code, file string, line int, pkg, fn string, args ...interface{}) *Frame {
	return &Frame{Err: err, Pkg: pkg, Func: fn, Args: args, Code: code, File: file, Line: line}
}

func (p *Frame) Error() string {
	return string(errorDetail(make([]byte, 0, 32), p))
}

func errorDetail(b []byte, p *Frame) []byte {
	if f, ok := p.Err.(*Frame); ok {
		b = errorDetail(b, f)
	} else {
		b = append(b, p.Err.Error()...)
		b = append(b, "\n\n===> errors stack:\n"...)
	}
	b = append(b, p.Pkg...)
	b = append(b, '.')
	b = append(b, p.Func...)
	b = append(b, '(')
	b = funcArgsDetail(b, p.Args)
	b = append(b, ")\n\t"...)
	b = append(b, p.File...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(p.Line), 10)
	b = append(b, ' ')
	b = append(b, p.Code...)
	b = append(b, '\n')
	return b
}

func funcArgsDetail(b []byte, args []interface{}) []byte {
	return b
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (p *Frame) Unwrap() error {
	return p.Err
}

// Format is required by fmt.Formatter
func (p *Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		io.WriteString(s, p.Error())
	case 's':
		io.WriteString(s, Err(p.Err).Error())
	case 'q':
		fmt.Fprintf(s, "%q", Err(p.Err).Error())
	}
}

// --------------------------------------------------------------------

// ErrorInfo is provided for backward compatibility
type ErrorInfo = Frame

// Detail is provided for backward compatibility
func (p *ErrorInfo) Detail(err error) *ErrorInfo {
	p.Code = err.Error()
	return p
}

// NestedObject is provided for backward compatibility
func (p *ErrorInfo) NestedObject() interface{} {
	return p.Err
}

// ErrorDetail is provided for backward compatibility
func (p *ErrorInfo) ErrorDetail() string {
	return p.Error()
}

// AppendErrorDetail is provided for backward compatibility
func (p *ErrorInfo) AppendErrorDetail(b []byte) []byte {
	return errorDetail(b, p)
}

// SummaryErr is provided for backward compatibility
func (p *ErrorInfo) SummaryErr() error {
	return p.Err
}

// Info is provided for backward compatibility
func Info(err error, cmd ...interface{}) *ErrorInfo {
	return &Frame{Err: err, Args: cmd}
}

// InfoEx is provided for backward compatibility
func InfoEx(calldepth int, err error, cmd ...interface{}) *ErrorInfo {
	return &Frame{Err: err, Args: cmd}
}

// Detail is provided for backward compatibility
func Detail(err error) string {
	return err.Error()
}

// --------------------------------------------------------------------
