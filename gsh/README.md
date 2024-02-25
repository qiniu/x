Go+ DevOps Tools
======

[![Language](https://img.shields.io/badge/language-Go+-blue.svg)](https://github.com/goplus/gop)
[![GitHub release](https://img.shields.io/github/v/tag/goplus/gop.svg?label=Go%2b+release)](https://github.com/goplus/gop/releases)
[![Discord](https://img.shields.io/badge/Discord-online-success.svg?logo=discord&logoColor=white)](https://discord.gg/mYjWCJDcAr)
[![GoDoc](https://pkg.go.dev/badge/github.com/qiniu/x/gsh.svg)](https://pkg.go.dev/github.com/qiniu/x/gsh)

This is an alternative to writing shell scripts.

Yes, now you can write `shell script` in Go+. It supports all shell commands.


## Usage

First, let's create a file named `./example.gsh` and write the following code:

```coffee
mkdir "testgsh"
```

You don't need a `go.mod` file, just enter `gop run ./example.gsh` directly to run.

It's strange to you that the file extension of Go+ source is not `.gop` but `.gsh`. It is only because Go+ register `.gsh` as a builtin [classfile](https://github.com/goplus/gop/blob/main/doc/classfile.md).

We can change `./example.gsh` more complicated:

```coffee
type file struct {
	name  string
	fsize int
}

mkdir! "testgsh"

mkdir "testgsh2"
lastErr!

mkdir "testgsh3"
if lastErr != nil {
	panic lastErr
}

capout => { ls }
println output.fields

capout => { ls "-l" }
files := [file{flds[8], flds[4].int!} for e <- output.split("\n") if flds := e.fields; flds.len > 2]
println files

rmdir "testgsh", "testgsh2", "testgsh3"
```


### Execute shell commands

There are many ways to execute shell commands. The simplest way is:

```coffee
mkdir "testgsh"
```

It is equivalent to:

```coffee
exec "mkdir", "testgsh"
```

or:

```coffee
exec "mkdir testgsh"
```

If a shell command is a Go/Go+ language keyword (eg. `go`), or the command is a relative or absolute path, you can only execute it in the latter two ways:

```coffee
exec "go", "version"
exec "./test.sh"
exec "/usr/bin/env gop run ."
```

You can also specify environment variables to run:

```coffee
exec "GOOS=linux GOARCH=amd64 go install ."
```


### Retrieve environment variables

You can get the value of an environment variable through `${XXX}`. For example:

```coffee
ls "${HOME}"
```


### Check last error

If we want to ensure `mkdir` successfully, there are three ways:

The simplest way is:

```coffee
mkdir! "testsh"  # will panic if mkdir failed
```

The second way is:

```coffee
mkdir "testsh"
lastErr!
```

Yes, `gsh` provides `lastErr` to check last error.

The third way is:

```coffee
mkdir "testsh"
if lastErr != nil {
    panic lastErr
}
```

This is the most familiar way to Go developers.


### Capture output of commands

And, `gsh` provides a way to capture output of commands:

```coffee
capout => {
    ...
}
```

Similar to `lastErr`, the captured output result is saved to `output`.

For example:

```coffee
capout => { ls "-l" }
println output
```

Here is a possible output:

```s
total 72
-rw-r--r--  1 xushiwei  staff  11357 Jun 19 00:20 LICENSE
-rw-r--r--  1 xushiwei  staff    127 Jun 19 10:00 README.md
-rw-r--r--  1 xushiwei  staff    365 Jun 19 00:25 example.gsh
-rw-r--r--  1 xushiwei  staff    126 Jun 19 09:33 go.mod
-rw-r--r--  1 xushiwei  staff    165 Jun 19 09:33 go.sum
-rw-r--r--  1 xushiwei  staff   1938 Jun 19 10:00 gop_autogen.go
```

We can use [Go+ powerful built-in data processing capabilities](https://github.com/goplus/gop/blob/main/doc/docs.md#data-processing) to process captured `output`:

```coffee
type file struct {
	name  string
	fsize int
}

files := [file{flds[8], flds[4].int!} for e <- output.split("\n") if flds := e.fields; flds.len > 2]
```

In this example, we split `output` by `"\n"`, and for each entry `e`, split it by spaces (`e.fields`) and save into `flds`. Condition `flds.len > 2` is to remove special line of output:

```s
total 72
```

At last, pick file name and size of all selected entries and save into `files`.
