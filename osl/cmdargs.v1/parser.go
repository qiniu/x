package cmdargs

import (
	"errors"
	"flag"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"qiniupkg.com/x/jsonutil.v7"
	"qiniupkg.com/x/log.v7"
	"qiniupkg.com/x/osl/cmdarg.v1"

	. "qiniupkg.com/dyn/vars.v1"
)

var (
	ErrParamsNotEnough = errors.New("params not enough")
	ErrTooMuchParams = errors.New("too much params")
	ErrUnsupportedFlagType = errors.New("unsupported flag type")
	ErrUnsupportedArgType = errors.New("unsupported argument type")
)

// ---------------------------------------------------------------------------

func hasOption(tag string, opt string) bool {

	for i := 0; i < len(tag); i++ {
		switch tag[i] {
		case ',':
			if strings.HasPrefix(tag[i+1:], opt) {
				tagLeft := tag[i+1+len(opt):]
				if tagLeft == "" || tagLeft[0] == ',' || tagLeft[0] == ' ' {
					return true
				}
			}
		case ' ':
			return false
		}
	}
	return false
}

func parseFlagArg(fv reflect.Value, f *flag.FlagSet, tag string) (err error) {

	n := 0
	for n < len(tag) {
		if tag[n] == ',' || tag[n] == ' ' {
			break
		}
		n++
	}

	name := tag[:n]
	usage := ""
	pos := strings.Index(tag[n:], " - ")
	if pos >= 0 {
		usage = tag[pos+3:]
	}

	switch fv.Kind() {
	case reflect.Ptr:
		switch fv.Elem().Kind() {
		case reflect.Bool:
			fv.Set(reflect.ValueOf(f.Bool(name, false, usage)))
		case reflect.Int:
			fv.Set(reflect.ValueOf(f.Int(name, 0, usage)))
		case reflect.Uint:
			fv.Set(reflect.ValueOf(f.Uint(name, 0, usage)))
		case reflect.Uint64:
			fv.Set(reflect.ValueOf(f.Uint64(name, 0, usage)))
		default:
			return ErrUnsupportedFlagType
		}
	case reflect.Bool:
		f.BoolVar(fv.Addr().Interface().(*bool), name, false, usage)
	case reflect.Int:
		f.IntVar(fv.Addr().Interface().(*int), name, 0, usage)
	case reflect.Uint:
		f.UintVar(fv.Addr().Interface().(*uint), name, 0, usage)
	case reflect.Uint64:
		f.Uint64Var(fv.Addr().Interface().(*uint64), name, 0, usage)
	default:
		return ErrUnsupportedFlagType
	}
	return nil
}

type parseArgOpts struct {
	ft       string
	keep     bool
	optional bool
}

func getFmttype(ft string, ftDefault int) int {

	switch ft {
	case "form": return Fmttype_Form
	case "text": return Fmttype_Text
	case "json": return Fmttype_Json
	}
	return ftDefault
}

func parseArgTag(tag string) (opts parseArgOpts, err error) {

	pos := strings.Index(tag, " - ")
	if pos >= 0 {
		tag = tag[:pos]
	}

	parts := strings.Split(tag, ",")
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "keep":
			opts.keep = true
		case "form", "text", "json":
			opts.ft = parts[i]
		case "opt":
			opts.optional = true
		default:
			err = errors.New("Unknown tag option: " + parts[i])
			return
		}
	}
	return
}

func parseArg(
	ctx *Context, fv reflect.Value, arg string, opts parseArgOpts) (err error) {

	kind := fv.Kind()
	switch kind {
	case reflect.String:
		if opts.keep { // 保留 $(var) 不要自动展开
			fv.SetString(arg)
			return
		}
		arg, err = ctx.SubstText(arg, getFmttype(opts.ft, Fmttype_Text))
		if err != nil {
			return
		}
		fv.SetString(arg)

	case reflect.Interface:
		argObj, err1 := cmdarg.Unmarshal(arg)
		if err1 != nil {
			return err1
		}
		if opts.keep { // 保留 $(var) 不要做 Subst
			fv.Set(reflect.ValueOf(argObj))
			return
		}
		argObj, err2 := ctx.Subst(argObj, getFmttype(opts.ft, Fmttype_Text))
		if err2 != nil {
			return err2
		}
		fv.Set(reflect.ValueOf(argObj))

	default:
		if kind >= reflect.Int && kind <= reflect.Int64 {
			arg, err = ctx.SubstText(arg, Fmttype_Text)
			if err != nil {
				return
			}
			intVal, err2 := strconv.ParseInt(arg, 10, 64)
			if err2 != nil {
				return err2
			}
			fv.SetInt(intVal)
			return nil
		}
		if kind >= reflect.Uint && kind <= reflect.Uintptr {
			arg, err = ctx.SubstText(arg, Fmttype_Text)
			if err != nil {
				return
			}
			uintVal, err2 := strconv.ParseUint(arg, 10, 64)
			if err2 != nil {
				return err2
			}
			fv.SetUint(uintVal)
			return nil
		}
		arg, err = ctx.SubstText(arg, getFmttype(opts.ft, Fmttype_Json))
		if err != nil {
			return
		}
		err = jsonutil.Unmarshal(arg, fv.Addr().Interface())
		if err != nil {
			log.Debug("parseCmdArgs failed:", err, "arg:", arg)
			return
		}
	}
	return nil
}

func parseVargs(
	ctx *Context, fv reflect.Value, args []string, opts parseArgOpts) (err error) {

	sliceType := fv.Type()
	n := len(args)
	sliceValue := reflect.MakeSlice(sliceType, n, n)
	for i, arg := range args {
		err = parseArg(ctx, sliceValue.Index(i), arg, opts)
		if err != nil {
			return
		}
	}
	fv.Set(sliceValue)
	return
}

func parseStructArgs(
	ctx *Context, strucType reflect.Type, cmd []string) (args reflect.Value, err error) {

	nField := strucType.NumField()

	hasFlag := false
	for i := 0; i < nField; i++ {
		sf := strucType.Field(i)
		if strings.HasPrefix(string(sf.Tag), "flag:") {
			hasFlag = true
			break
		}
	}

	args = reflect.New(strucType)
	argsRef := args.Elem()

	if hasFlag {
		f := flag.NewFlagSet(cmd[0], 0)
		for i := 0; i < nField; i++ {
			sf := strucType.Field(i)
			if strings.HasPrefix(string(sf.Tag), "f") {
				err = parseFlagArg(argsRef.Field(i), f, sf.Tag.Get("flag"))
				if err != nil {
					return
				}
			}
		}
		err = f.Parse(cmd[1:])
		if err != nil {
			return
		}
		cmd = f.Args()
	} else {
		cmd = cmd[1:]
	}

	icmd := 0
	for i := 0; i < nField; i++ {
		sf := strucType.Field(i)
		if strings.HasPrefix(string(sf.Tag), "arg:") {
			tag := sf.Tag.Get("arg")
			opts, err2 := parseArgTag(tag)
			if err2 != nil {
				err = err2
				return
			}
			fv := argsRef.Field(i)
			if fv.Kind() == reflect.Slice { // 不定参数
				err = parseVargs(ctx, fv, cmd[icmd:], opts)
				return
			}
			if icmd >= len(cmd) {
				if opts.optional { // 可选参数
					return
				}
				err = ErrParamsNotEnough
				return
			}
			err = parseArg(ctx, fv, cmd[icmd], opts)
			if err != nil {
				return
			}
			icmd++
		}
	}
	if icmd != len(cmd) {
		err = ErrTooMuchParams
	}
	return
}

func Parse(
	ctx *Context, argsType reflect.Type, cmd []string) (args reflect.Value, err error) {

	switch argsType.Kind() {
	case reflect.Ptr: // may be args *xxxArgs
		strucType := argsType.Elem()
		if strucType.Kind() == reflect.Struct {
			return parseStructArgs(ctx, strucType, cmd)
		}
	case reflect.Slice: // may be args []string
		if argsType.Elem().Kind() == reflect.String {
			return reflect.ValueOf(cmd), nil
		}
	}
	err = syscall.EINVAL
	return
}

// ---------------------------------------------------------------------------

