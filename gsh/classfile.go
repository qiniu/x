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
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
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

// InitApp initializes an App instance. It is provided so that an App instance
// embedded in a struct can be initialized from another package.
func InitApp(app *App) {
	app.initApp()
}

func (p *App) initApp() {
	p.fin = os.Stdin
	p.fout = os.Stdout
	p.ferr = os.Stderr
}

// Gop_Env retrieves the value of the environment variable named by the key.
func (p *App) Gop_Env(key string) string {
	return Sys.Getenv(key)
}

func (p *App) execWith(env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = p.fin
	cmd.Stdout = p.fout
	cmd.Stderr = p.ferr
	cmd.Env = env
	p.err = Sys.Run(cmd)
	return p.err
}

// Gop_Exec executes a shell command.
func (p *App) Gop_Exec(name string, args ...string) error {
	return p.execWith(nil, name, args...)
}

// Exec executes a shell command with specified environs.
func (p *App) Exec__0(env map[string]string, name string, args ...string) error {
	var cmdEnv []string
	if len(env) > 0 {
		cmdEnv = Setenv__0(Sys.Environ(), env)
	}
	return p.execWith(cmdEnv, name, args...)
}

// Exec executes a shell command line with $env variables support.
//   - exec "XGO_GOCMD=tinygo xgo run ."
//   - exec "ls -l $HOME"
func (p *App) Exec__1(cmdline string) error {
	var iCmd = -1
	var items = strings.Fields(cmdline)
	for i, e := range items {
		pos := strings.IndexAny(e, "=$")
		if pos >= 0 && e[pos] == '=' {
			if strings.IndexByte(e[pos+1:], '$') >= 0 {
				items[i] = Sys.ExpandEnv(e)
			}
			continue
		}
		if iCmd < 0 {
			iCmd = i
		}
		if pos >= 0 {
			items[i] = Sys.ExpandEnv(e)
		}
	}
	if iCmd < 0 {
		return errors.New("exec: no command")
	}
	var env []string
	if iCmd > 0 {
		env = Setenv__2(Sys.Environ(), items[:iCmd])
	}
	return p.execWith(env, items[iCmd], items[iCmd+1:]...)
}

// Exec executes a shell command.
func (p *App) Exec__2(name string, args ...string) error {
	return p.execWith(nil, name, args...)
}

// LastErr returns error of last command execution.
func (p *App) LastErr() error {
	return p.err
}

// ExitCode returns exit code of last command execution. Bash-scripting exit codes:
// 1: Catchall for general errors
// 2: Misuse of shell builtins (according to Bash documentation)
// 126: Command invoked cannot execute
// 127: Command not found
// 128+n: Fatal error signal "n"
// 254: Unknown error(*)
func (p *App) ExitCode() int {
	if p.err == nil {
		return 0
	}
	switch e := p.err.(type) {
	case *exec.ExitError:
		return e.ProcessState.ExitCode()
	default:
		switch e.Error() {
		case "exec: no command":
			return 127
		case "exec: not started":
			return 126
		default:
			return 254
		}
	}
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
