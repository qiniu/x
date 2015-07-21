package cmdarg

import (
	"strings"

	"qiniupkg.com/dyn/jsonext.v1"
)

// ---------------------------------------------------------------------------

func UnmarshalText(text string) (v interface{}, err error) {

	if strings.HasPrefix(text, "$(") {
		err = jsonext.UnmarshalString(text, &v)
		return
	}
	return text, nil
}

func Unmarshal(text string) (v interface{}, err error) {

	if len(text) == 0 {
		return "", nil
	}

	c := text[0] // true, false, null
	if c <= 33 || (c >= 'A' && c <= 'Z') || c == '/' {
		return text, nil
	}

	if c >= 'a' && c <= 'z' {
		switch text {
		case "true": return true, nil
		case "false": return false, nil
		case "null": return nil, nil
		default:
			return text, nil
		}
	}

	err = jsonext.UnmarshalString(text, &v)
	return
}

// ---------------------------------------------------------------------------

