/*
 Copyright 2024 Qiniu Limited (qiniu.com)

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

package gsh

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

const (
	GopPackage = true
)

// App is project class of this classfile.
type App struct {
	fout io.Writer
	ferr io.Writer
	fin  io.Reader
	cout string
	err  error
}

func (p *App) initApp() {
	p.fin = os.Stdin
	p.fout = os.Stdout
	p.ferr = os.Stderr
}

// Gop_Exec executes a shell command.
func (p *App) Gop_Exec(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = p.fin
	cmd.Stdout = p.fout
	cmd.Stderr = p.ferr
	p.err = cmd.Run()
	return p.err
}

// LastErr returns error of last command execution.
func (p *App) LastErr() error {
	return p.err
}

// Capout captures stdout of doSth() execution and save it to output.
func (p *App) Capout(doSth func()) (string, error) {
	var out bytes.Buffer
	old := p.fout
	p.fout = &out
	defer func() {
		p.fout = old
	}()
	doSth()
	p.cout = out.String()
	return p.cout, p.err
}

// Output returns result of last capout.
func (p *App) Output() string {
	return p.cout
}

// Gopt_App_Main is main entry of this classfile.
func Gopt_App_Main(a interface{ initApp() }) {
	a.initApp()
	a.(interface{ MainEntry() }).MainEntry()
}
