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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/qiniu/x/cmdjsonl"
)

type Context struct {
	vars map[string]float64
	line int
}

func (ctx *Context) resolve(p any) float64 {
	switch v := p.(type) {
	case float64:
		return v
	case string:
		if strings.HasPrefix(v, "$") {
			name := v[1:]
			if val, ok := ctx.vars[name]; ok {
				return val
			}
			panic(fmt.Sprintf("undefined variable: %s", name))
		}
	}
	panic(fmt.Sprintf("invalid operand: %v", p))
}

func (ctx *Context) store(result float64) {
	ctx.line++
	name := fmt.Sprintf("out%d", ctx.line)
	ctx.vars[name] = result
	fmt.Printf("%s = %v\n", name, result)
}

type BinOp struct {
	A any `json:"a"`
	B any `json:"b"`
}

func main() {
	ctx := &Context{vars: make(map[string]float64)}

	var p cmdjsonl.Parser
	p.HandleFunc("add", func(op BinOp) {
		a := ctx.resolve(op.A)
		b := ctx.resolve(op.B)
		ctx.store(a + b)
	})

	p.HandleFunc("sub", func(op BinOp) {
		a := ctx.resolve(op.A)
		b := ctx.resolve(op.B)
		ctx.store(a - b)
	})

	p.HandleFunc("mul", func(op BinOp) {
		a := ctx.resolve(op.A)
		b := ctx.resolve(op.B)
		ctx.store(a * b)
	})

	p.HandleFunc("quo", func(op BinOp) {
		a := ctx.resolve(op.A)
		b := ctx.resolve(op.B)
		if b == 0 {
			panic("division by zero")
		}
		ctx.store(a / b)
	})

	err := p.Parse(os.Stdin, 1024)
	if err != nil {
		panic(err)
	}
}
