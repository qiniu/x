# cmdjsonl

[![Go Reference](https://pkg.go.dev/badge/github.com/qiniu/x/cmdjsonl.svg)](https://pkg.go.dev/github.com/qiniu/x/cmdjsonl)

`cmdjsonl` is a Go package for parsing data in the `application/x-cmdjsonl` format. This format represents command-style invocations as line-oriented text, making it convenient for streaming, logging, and cross-language/cross-process communication.

## Data Format

The `application/x-cmdjsonl` format is defined as follows:

- UTF-8 encoding;
- Each line consists of a `Command` followed by a space and a JSON object, with lines separated by a newline character (`\n`).

Example:

```
add {"a":1,"b":2}
greet {"name":"Alice"}
log {"level":"info","msg":"hello world"}
```

## Installation

```bash
go get github.com/qiniu/x/cmdjsonl
```

## Quick Start

```go
package main

import (
	"strings"

	"github.com/qiniu/x/cmdjsonl"
)

type AddParam struct {
	A int `json:"a"`
	B int `json:"b"`
}

type GreetParam struct {
	Name string `json:"name"`
}

func main() {
	var p cmdjsonl.Parser

	// Register a handler with a non-pointer parameter
	p.HandleFunc("add", func(param AddParam) {
		// ...
	})

	// Register a handler with a pointer parameter that returns an error
	p.HandleFunc("greet", func(param *GreetParam) error {
		// ...
		return nil
	})

	data := `add {"a":1,"b":2}
greet {"name":"Alice"}
`
	if err := p.Parse(strings.NewReader(data), 4096); err != nil {
		panic(err)
	}
}
```

## API

### type Parser

```go
type Parser struct {
	// ...
}
```

`Parser` parses data in the `cmdjsonl` format and dispatches the parsed commands to registered handler functions.

#### func (*Parser) HandleFunc

```go
func (p *Parser) HandleFunc(cmd string, fn any) error
```

Registers a handler function `fn` for the specified command `cmd`.

The handler function `fn` must satisfy the following requirements:

- It must be a function;
- It must have exactly one parameter, which can be either a pointer or a non-pointer type;
- It must return zero or one value; if it returns a value, that value's type must be `error`.
If `fn` does not meet these requirements, `HandleFunc` returns an `*InvalidHandler` error.

Examples of parameter types:

```go
// Non-pointer parameter
p.HandleFunc("foo", func(param FooParam) { ... })

// Pointer parameter
p.HandleFunc("bar", func(param *BarParam) error { ... })

// No return value
p.HandleFunc("baz", func(param BazParam) { ... })
```

#### func (*Parser) Parse

```go
func (p *Parser) Parse(in io.Reader, maxLineSize int) error
```

Reads and parses data line by line from `in`:

- `maxLineSize` specifies the maximum buffer size (in bytes) for a single line; if a line exceeds this length, an error is returned;
- Each line is split at the first space into a `command` and a `JSON payload`;
- The handler registered for the command is looked up, and the JSON payload is unmarshaled into that handler's parameter type;
- The handler function is then invoked; if it returns a non-`nil` `error`, parsing stops and that error is returned;
- When `io.EOF` is reached, parsing completes normally and `nil` is returned.
### type InvalidHandler

```go
type InvalidHandler struct {
	Cmd    string // The command for which the handler is invalid.
	Reason string // The reason why the handler is invalid.
}

func (e *InvalidHandler) Error() string
```

Represents an error returned by `HandleFunc` when a registered handler function does not meet the requirements.

### type ParseError

```go
type ParseError struct {
	Line int    // The line number where the error occurred (1-indexed).
	When string // The stage at which the error occurred.
	Msg  string // The error message.
}

func (e *ParseError) Error() string
```

Represents an error that occurred during `Parse`, including the line number, the stage (e.g., `ReadLine`, `ParseCommand`, `UnmarshalParam`, `CallHandler <cmd>`), and a message describing the error, making it easy to pinpoint the issue.

## Error Handling

`Parse` stops and returns a `*ParseError` in the following cases:

| When             | Trigger Condition                                              |
| ---------------- | --------------------------------------------------------------- |
| `ReadLine`        | An I/O error occurred while reading a line, or a line exceeds `maxLineSize` |
| `ParseCommand`    | No space found in the line, or the command has no registered handler |
| `UnmarshalParam`  | Failed to unmarshal the JSON payload                            |
| `CallHandler <cmd>` | The handler function returned a non-`nil` `error`             |

When the end of input (`io.EOF`) is reached normally, `Parse` returns `nil`, indicating successful completion.
