package rpc

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	. "golang.org/x/net/context"
)

var UserAgent = "Golang qiniu/rpc package"

// --------------------------------------------------------------------

type key int // key is unexported and used for Context

const (
	xlogKey key = 0
)

type xlogger interface {
	ReqId() string
	Xput(logs []string)
}

func NewXlogContext(ctx Context, l xlogger) Context {
	return WithValue(ctx, xlogKey, l)
}

func XlogFromContext(ctx Context) (l xlogger, ok bool) {
	l, ok = ctx.Value(xlogKey).(xlogger)
	return
}

// --------------------------------------------------------------------

type Client struct {
	*http.Client
}

var DefaultClient = Client{&http.Client{Transport: http.DefaultTransport}}

// --------------------------------------------------------------------

func (r Client) DoRequest(ctx Context, method, url string) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	return r.Do(ctx, req)
}

func (r Client) DoRequestWith(
	ctx Context, method, url1 string,
	bodyType string, body io.Reader, bodyLength int) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url1, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(bodyLength)
	return r.Do(ctx, req)
}

func (r Client) DoRequestWith64(
	ctx Context, method, url1 string,
	bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url1, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = bodyLength
	return r.Do(ctx, req)
}

func (r Client) DoRequestWithForm(
	ctx Context, method, url1 string, data map[string][]string) (resp *http.Response, err error) {

	msg := url.Values(data).Encode()
	if method == "GET" || method == "HEAD" || method == "DELETE" {
		if strings.ContainsRune(url1, '?') {
			url1 += "&"
		} else {
			url1 += "?"
		}
		return r.DoRequest(ctx, method, url1 + msg)
	}
	return r.DoRequestWith(
		ctx, method, url1, "application/x-www-form-urlencoded", strings.NewReader(msg), len(msg))
}

func (r Client) DoRequestWithJson(
	ctx Context, method, url1 string, data interface{}) (resp *http.Response, err error) {

	msg, err := json.Marshal(data)
	if err != nil {
		return
	}
	return r.DoRequestWith(
		ctx, method, url1, "application/json", bytes.NewReader(msg), len(msg))
}

func (r Client) Do(ctx Context, req *http.Request) (resp *http.Response, err error) {

	if ctx == nil {
		ctx = Background()
	}

	l, lok := XlogFromContext(ctx)
	if lok {
		req.Header.Set("X-Reqid", l.ReqId())
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", UserAgent)
	}

	transport := r.Transport // don't change r.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// avoid cancel() is called before Do(req), but isn't accurate
	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	default:
	}

	if tr, ok := getRequestCanceler(transport); ok { // support CancelRequest
		reqC := make(chan bool, 1)
		go func() {
			resp, err = r.Client.Do(req)
			reqC <- true
		}()
		select {
		case <-reqC:
		case <-ctx.Done():
			tr.CancelRequest(req)
			<-reqC
			err = ctx.Err()
		}
	} else {
		resp, err = r.Client.Do(req)
	}
	if err != nil {
		return
	}

	if lok {
		details := resp.Header["X-Log"]
		if len(details) > 0 {
			l.Xput(details)
		}
	}
	return
}

// --------------------------------------------------------------------

type errorInfo struct {
	Err   string `json:"error"`
	Key   string `json:"key,omitempty"`
	Errno int    `json:"errno,omitempty"`
	Code  int    `json:"code"`
	Reqid string `json:"reqid"`
}

func (r *errorInfo) ErrorDetail() string {
	msg, _ := json.Marshal(r)
	return string(msg)
}

func (r *errorInfo) Error() string {
	if r.Err != "" {
		return r.Err
	}
	return http.StatusText(r.Code)
}

func (r *errorInfo) HttpCode() int {
	return r.Code
}

// --------------------------------------------------------------------

func parseError(e *errorInfo, r io.Reader) {

	body, err1 := ioutil.ReadAll(r)
	if err1 != nil {
		e.Err = err1.Error()
		return
	}

	var ret struct {
		Err   string `json:"error"`
		Key   string `json:"key"`
		Errno int    `json:"errno"`
	}
	if json.Unmarshal(body, &ret) == nil && ret.Err != "" {
		// qiniu error msg style returns here
		e.Err, e.Key, e.Errno = ret.Err, ret.Key, ret.Errno
		return
	}
	e.Err = string(body)
}

func ResponseError(resp *http.Response) (err error) {

	e := &errorInfo{
		Reqid: resp.Header.Get("X-Reqid"),
		Code:  resp.StatusCode,
	}
	if resp.StatusCode > 299 {
		if resp.ContentLength != 0 {
			ct, ok := resp.Header["Content-Type"]
			if ok && strings.HasPrefix(ct[0], "application/json") {
				parseError(e, resp.Body)
			}
		}
	}
	return e
}

func CallRet(ctx Context, ret interface{}, resp *http.Response) (err error) {

	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				return
			}
		}
		if resp.StatusCode == 200 {
			return nil
		}
	}
	return ResponseError(resp)
}

func (r Client) CallWithForm(
	ctx Context, ret interface{}, method, url1 string, param map[string][]string) (err error) {

	resp, err := r.DoRequestWithForm(ctx, method, url1, param)
	if err != nil {
		return err
	}
	return CallRet(ctx, ret, resp)
}

func (r Client) CallWithJson(
	ctx Context, ret interface{}, method, url1 string, param interface{}) (err error) {

	resp, err := r.DoRequestWithJson(ctx, method, url1, param)
	if err != nil {
		return err
	}
	return CallRet(ctx, ret, resp)
}

func (r Client) CallWith(
	ctx Context, ret interface{}, method, url1, bodyType string, body io.Reader, bodyLength int) (err error) {

	resp, err := r.DoRequestWith(ctx, method, url1, bodyType, body, bodyLength)
	if err != nil {
		return err
	}
	return CallRet(ctx, ret, resp)
}

func (r Client) CallWith64(
	ctx Context, ret interface{}, method, url1, bodyType string, body io.Reader, bodyLength int64) (err error) {

	resp, err := r.DoRequestWith64(ctx, method, url1, bodyType, body, bodyLength)
	if err != nil {
		return err
	}
	return CallRet(ctx, ret, resp)
}

func (r Client) Call(
	ctx Context, ret interface{}, method, url1 string) (err error) {

	resp, err := r.DoRequestWith(ctx, method, url1, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return err
	}
	return CallRet(ctx, ret, resp)
}

// ---------------------------------------------------------------------------

type requestCanceler interface {
	CancelRequest(req *http.Request)
}

func getRequestCanceler(tp http.RoundTripper) (rc requestCanceler, ok bool) {

	v := reflect.ValueOf(tp)

subfield:
	// panic if the Field is unexported (but this can be detected in developing)
	if rc, ok = v.Interface().(requestCanceler); ok {
		return
	}
	v = reflect.Indirect(v)
	if v.Kind() == reflect.Struct {
		for i := v.NumField() - 1; i >= 0; i-- {
			sv := v.Field(i)
			if sv.Kind() == reflect.Interface {
				sv = sv.Elem()
			}
			if sv.MethodByName("RoundTrip").IsValid() {
				v = sv
				goto subfield
			}
		}
	}
	return
}

// --------------------------------------------------------------------

