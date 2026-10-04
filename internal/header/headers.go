package header

import (
	"bytes"
	"fmt"
	"strings"
)

var ErrorInvalidHeader = fmt.Errorf("error invalid header")
const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
    return make(Headers)
}

func validHeaderName(s string) bool {
	return len(s) > 0 && strings.IndexFunc(s, func(r rune) bool { 
		return !(
			('A' <= r  && r <= 'Z') || 
			('a' <= r  && r <= 'z') || 
			('0' <= r  && r <= '9') || 
			(r == '!' || r == '#' || r == '$' || r == '%' || r == '&') ||
			(r == '\'' || r == '*' || r == '+' || r == '-' || r == '.') ||
			(r == '^' || r == '_' || r == '`' || r == '|' || r == '~'))
	}) == -1
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	clrfIdx := bytes.Index(data, []byte(crlf))
	if clrfIdx == -1 {
		return 0, false, nil
	}
	if clrfIdx == 0 {
		return 2, true, nil
	}
	line := string(data[:clrfIdx])
	before, after, ok := strings.Cut(line, ":")
	if !ok {
		return 0, false, ErrorInvalidHeader
	}
	if before != strings.TrimLeft(before, " ") || before != strings.TrimRight(before, " ") {
		return 0, false, ErrorInvalidHeader 
	}
	if !validHeaderName(before) {
		return 0, false, ErrorInvalidHeader
	}
	name := strings.ToLower(before)
	value := strings.TrimSpace(after)
	numBytesParsed := len(line) + 2
	if _, ok := h[name]; !ok {
		h[name] = value
	} else {
		h[name] += ", " + value
	}
	return numBytesParsed, false, nil
}