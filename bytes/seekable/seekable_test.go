package seekable

import (
	"bytes"
	"net/http"
	"testing"
)

func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Fatal("assertEqual:", a, b)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal("assertNoError:", err)
	}
}

func TestSeekable_EOFIfReqAlreadyParsed(t *testing.T) {
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assertNoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	req.ParseForm()
	_, err = New(req)
	assertEqual(t, err.Error(), "EOF")
}

func TestSeekable_WorkaroundForEOF(t *testing.T) {
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assertNoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, _ = New(req)
	req.ParseForm()
	assertEqual(t, req.FormValue("a"), "1")
	_, err = New(req)
	assertNoError(t, err)
}

func TestSeekable(t *testing.T) {
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	assertNoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, err = New(req)
	assertNoError(t, err)
}
