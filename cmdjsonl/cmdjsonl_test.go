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
	"errors"
	"strings"
	"testing"
)

func TestHandleFunc(t *testing.T) {
	var p Parser
	err := p.HandleFunc("test", 100)
	if err == nil || err.Error() != "invalid handler for command 'test': handler must be a function" {
		t.Error("HandleFunc:", err)
	}
	err = p.HandleFunc("test", func(a, b int) {})
	if err == nil || err.Error() != "invalid handler for command 'test': handler must have exactly one parameter" {
		t.Error("HandleFunc:", err)
	}
	err = p.HandleFunc("test", func(a int) (int, error) { return 0, nil })
	if err == nil || err.Error() != "invalid handler for command 'test': handler must return at most one value, which must be of type error" {
		t.Error("HandleFunc:", err)
	}
	err = p.HandleFunc("test", func(a int) error { return nil })
	if err != nil {
		t.Error("HandleFunc:", err)
	}
}

func parse(p Parser, text string) error {
	return p.Parse(strings.NewReader(text), 64)
}

func TestParse(t *testing.T) {
	type TestCmd struct {
		Name string
		Age  int
	}
	var p Parser
	p.HandleFunc("test", func(cmd *TestCmd) error {
		if cmd.Age == 30 {
			return nil
		}
		return errors.New("handler error")
	})
	p.HandleFunc("testNonPtr", func(cmd TestCmd) error {
		return nil
	})
	err := parse(p, `testNonPtr {"Name": "Alice", "Age": 20}
test {"Name": "Ken", "Age": 30}
`)
	if err != nil {
		t.Error("Parse:", err)
	}
	err = parse(p, `test {"Name": "Alice", "Age": 20}`)
	if err == nil || err.Error() != "1: parse error when CallHandler test: handler error" {
		t.Error("Parse:", err)
	}
	err = parse(p, `test {"Name": "Alice",`)
	if err == nil || err.Error() != "1: parse error when UnmarshalParam: unexpected end of JSON input" {
		t.Error("Parse:", err)
	}
	err = parse(p, `abc {"Name": "Bob", "Age": "30"}`)
	if err == nil || err.Error() != "1: parse error when ParseCommand: unknown command 'abc'" {
		t.Error("Parse:", err)
	}
	err = parse(p, `abc`)
	if err == nil || err.Error() != "1: parse error when ParseCommand: no space found" {
		t.Error("Parse:", err)
	}
	err = parse(p, strings.Repeat("abc", 100))
	if err == nil || err.Error() != "1: parse error when ReadLine: line too long" {
		t.Error("Parse:", err)
	}
}
