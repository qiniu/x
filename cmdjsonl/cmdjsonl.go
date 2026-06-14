/*
 Copyright 2026 Qiniu Limited (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmdjsonl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
)

/* ---------------------------------------------------------------------------

MIEM-Type: application/x-cmdjsonl

Data format:

* UTF-8 Encoding
* Each line is a Command + " " + JSON object, separated by a newline character (`\n`).

// -------------------------------------------------------------------------*/

type handler struct {
	fn     reflect.Value // The handler function to be called when the command is invoked.
	param  reflect.Type  // The type of the parameter that the handler function accepts.
	nonPtr bool          // Indicates whether the parameter type is non-pointer.
	retErr bool          // Indicates whether the handler function returns an error.
}

// Parser parses input data in the cmdjsonl format (MIME type: application/x-cmdjsonl) and
// dispatches commands to registered handlers. Each line of input is expected to contain a
// command followed by a JSON object, separated by a space. The parser reads the input line
// by line, extracts the command and the JSON object, and invokes the corresponding handler.
type Parser struct {
	cmds map[string]handler
}

type InvalidHandler struct {
	Cmd    string // The command for which the handler is invalid.
	Reason string // The reason why the handler is invalid.
}

func (e *InvalidHandler) Error() string {
	return "invalid handler for command '" + e.Cmd + "': " + e.Reason
}

// HandleFunc registers a handler function for a specific command. The handler function
// must have exactly one parameter, which can be either a pointer or a non-pointer type.
// The function can optionally return an error. If the handler function does not meet
// these requirements, an InvalidHandler error is returned.
func (p *Parser) HandleFunc(cmd string, fn any) error {
	var h handler
	h.fn = reflect.ValueOf(fn)
	if h.fn.Kind() != reflect.Func {
		return &InvalidHandler{Cmd: cmd, Reason: "handler must be a function"}
	}
	tfn := h.fn.Type()
	if tfn.NumIn() != 1 {
		return &InvalidHandler{Cmd: cmd, Reason: "handler must have exactly one parameter"}
	}
	if nout := tfn.NumOut(); nout > 0 {
		if nout != 1 || tfn.Out(0) != reflect.TypeFor[error]() {
			return &InvalidHandler{Cmd: cmd, Reason: "handler must return at most one value, which must be of type error"}
		}
		h.retErr = true
	}
	if p.cmds == nil {
		p.cmds = make(map[string]handler)
	}
	h.param = tfn.In(0)
	if h.param.Kind() == reflect.Pointer {
		h.param = h.param.Elem()
	} else {
		h.nonPtr = true
	}
	p.cmds[cmd] = h
	return nil
}

// ParseError represents an error that occurred during parsing of the input. It includes
// the line number where the error occurred, a description of when the error happened, and
// a message describing the error itself.
type ParseError struct {
	Line int
	When string
	Msg  string
}

func (p *ParseError) Error() string {
	return strconv.Itoa(p.Line) + ": parse error when " + p.When + ": " + p.Msg
}

// Parse reads from the provided io.Reader line by line, expecting each line to contain
// a command followed by a JSON object. It uses the registered handlers to process each
// command. If any error occurs during reading, parsing, or handling a command, a ParseError
// is returned with details about the error.
func (p *Parser) Parse(in io.Reader, maxLine int) error {
	lnum := 0
	r := bufio.NewReaderSize(in, maxLine)
	for {
		line, isPrefix, err := r.ReadLine()
		lnum++
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return &ParseError{Line: lnum, When: "ReadLine", Msg: err.Error()}
		}
		if isPrefix {
			return &ParseError{Line: lnum, When: "ReadLine", Msg: "line too long"}
		}
		pos := bytes.IndexByte(line, ' ')
		if pos < 0 {
			return &ParseError{Line: lnum, When: "ParseCommand", Msg: "no space found"}
		}
		cmd := string(line[:pos])
		h, ok := p.cmds[cmd]
		if !ok {
			return &ParseError{Line: lnum, When: "ParseCommand", Msg: "unknown command '" + cmd + "'"}
		}
		param := reflect.New(h.param)
		err = json.Unmarshal(line[pos+1:], param.Interface())
		if err != nil {
			return &ParseError{Line: lnum, When: "UnmarshalParam", Msg: err.Error()}
		}
		if h.nonPtr {
			param = param.Elem()
		}
		ret := h.fn.Call([]reflect.Value{param})
		if h.retErr {
			vErr := ret[0]
			if !vErr.IsNil() {
				return &ParseError{Line: lnum, When: "CallHandler " + cmd, Msg: vErr.Interface().(error).Error()}
			}
		}
	}
}

// ---------------------------------------------------------------------------
