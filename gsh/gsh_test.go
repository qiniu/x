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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type mockOS struct{}

func (p mockOS) Environ() []string {
	return append(make([]string, 0, len(mockEnv)), mockEnv...)
}

func (p mockOS) ExpandEnv(s string) string {
	return os.Expand(s, func(name string) string {
		return Getenv(mockEnv, name)
	})
}

func (p mockOS) Getenv(key string) string {
	return Getenv(mockEnv, key)
}

func (p mockOS) Run(c *exec.Cmd) error {
	if mockEcho {
		fmt.Fprintln(c.Stdout, c.Env, c.Args)
	}
	return mockRunErr
}

var (
	mockEnv    = []string{"FOO=foo", "BAR=bar"}
	mockRunErr error
	mockEcho   bool
)

func init() {
	Sys = mockOS{}
}

func capout(app *App, doSth func()) (ret string, err error) {
	mockEcho = true
	ret, err = app.Capout(doSth)
	mockEcho = false
	return
}

func lasterr(app *App, err error) {
	mockRunErr = err
	app.XGo_Exec("ls")
	mockRunErr = nil
}

// -----------------------------------------------------------

type M map[string]string

type myApp struct {
	App
	t *testing.T
}

func (p *myApp) MainEntry() {
	t := p.t
	err := p.XGo_Exec("ls", "-l")
	check(t, err)
	err = p.Exec__2("ls", "-l")
	check(t, err)
}

func TestMainEntry(t *testing.T) {
	XGot_App_Main(&myApp{t: t})
}

func TestOS(t *testing.T) {
	var sys defaultOS
	sys.Environ()
	sys.ExpandEnv("foo")
	sys.Getenv("foo")
	sys.Run(new(exec.Cmd))
}

func TestEnv(t *testing.T) {
	if Getenv(nil, "foo") != "" {
		t.Fatal("TestEnv: Getenv")
	}
	if ret := Setenv__1(nil, "k", "v"); len(ret) != 1 || ret[0] != "k=v" {
		t.Fatal("TestEnv Setenv:", ret)
	}
}

func TestExecWithEnv(t *testing.T) {
	var app App
	app.initApp()
	if v := app.XGo_Env("BAR"); v != "bar" {
		t.Fatal("app.XGo_Env:", v)
	}
	capout(&app, func() {
		err := app.Exec__0(M{"FOO": "123"}, "./app", "$FOO")
		check(t, err)
	})
	if v := app.Output(); v != "[FOO=123 BAR=bar] [./app $FOO]\n" {
		t.Fatal("TestExecWithEnv:", v)
	}
}

func TestExecSh(t *testing.T) {
	var app App
	app.initApp()
	capout(&app, func() {
		err := app.Exec__1("FOO=123 ./app $BAR")
		check(t, err)
	})
	if v := app.Output(); v != "[FOO=123 BAR=bar] [./app bar]\n" {
		t.Fatal("TestExecSh:", v)
	}
}

func TestExecSh2(t *testing.T) {
	var app App
	app.initApp()
	capout(&app, func() {
		err := app.Exec__1("FOO=$BAR ./app $FOO")
		check(t, err)
	})
	if v := app.Output(); v != "[FOO=bar BAR=bar] [./app foo]\n" {
		t.Fatal("TestExecSh2:", v)
	}
}

func TestExecSh2_Env(t *testing.T) {
	var app App
	app.initApp()
	capout(&app, func() {
		err := app.Exec__1("FOO=$BAR ./app ${FOO}")
		check(t, err)
	})
	if v := app.Output(); v != "[FOO=bar BAR=bar] [./app foo]\n" {
		t.Fatal("TestExecSh2:", v)
	}
}

func TestExecSh3(t *testing.T) {
	var app App
	app.initApp()
	err := app.Exec__1("FOO=$BAR X=1")
	checkErr(t, err, "exec: no command")
}

func TestExecSh4(t *testing.T) {
	var app App
	app.initApp()
	capout(&app, func() {
		err := app.Exec__1("FOO=$BAR X=1 ./app")
		check(t, err)
	})
	if v := app.Output(); v != "[FOO=bar BAR=bar X=1] [./app]\n" {
		t.Fatal("TestExecSh4:", v)
	}
}

func TestExitCode(t *testing.T) {
	var app App
	app.initApp()
	lasterr(&app, nil)
	check(t, app.LastErr())
	if v := app.ExitCode(); v != 0 {
		t.Fatal("ExitCode:", v)
	}
	lasterr(&app, errors.New("exec: no command"))
	if v := app.ExitCode(); v != 127 {
		t.Fatal("ExitCode:", v)
	}
	lasterr(&app, errors.New("exec: not started"))
	if v := app.ExitCode(); v != 126 {
		t.Fatal("ExitCode:", v)
	}
	lasterr(&app, errors.New("unknown"))
	if v := app.ExitCode(); v != 254 {
		t.Fatal("ExitCode:", v)
	}
	lasterr(&app, new(exec.ExitError))
	if v := app.ExitCode(); v != -1 {
		t.Fatal("ExitCode:", v)
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func checkErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil || err.Error() != msg {
		t.Fatal(err)
	}
}

// -----------------------------------------------------------
