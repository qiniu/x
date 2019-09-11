package jsonutil

import (
	"testing"
)

func Test(t *testing.T) {

	var ret struct {
		ID string `json:"id"`
	}
	err := Unmarshal(`{"id": "123"}`, &ret)
	if err != nil {
		t.Fatal("Unmarshal failed:", err)
	}
	if ret.ID != "123" {
		t.Fatal("Unmarshal uncorrect:", ret.ID)
	}
}
