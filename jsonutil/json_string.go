package jsonutil

import (
	"encoding/json"
	"reflect"
	"unsafe"
)

// ----------------------------------------------------------

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data string, v interface{}) error {

	sh := *(*reflect.StringHeader)(unsafe.Pointer(&data))
	arr := (*[1 << 30]byte)(unsafe.Pointer(sh.Data))
	return json.Unmarshal(arr[:sh.Len], v)
}

// ----------------------------------------------------------

// Stringify converts a value into string.
func Stringify(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ----------------------------------------------------------
