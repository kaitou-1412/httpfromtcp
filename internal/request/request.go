package request

import (
	"fmt"
	"io"
	"strings"
)

const crlf = "\r\n"
const bufferSize = 8

type ParserState string 

const (
	ParserStateInitialized ParserState = "initialized"
	ParserStateDone ParserState = "done"
)  
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
    RequestLine RequestLine
	State ParserState
}

func isOnlyUppercase(s string) bool {
    return len(s) > 0 && strings.IndexFunc(s, func(r rune) bool { return r < 'A' || r > 'Z' }) == -1
}

func parseRequestLine(data string) (*RequestLine, int, error) {
	hasCRLF := strings.Contains(data, "\r\n")
	if !hasCRLF {
		return nil, 0, nil
	}
	lines := strings.Split(data, "\r\n")
	if len(lines) < 1 {
		return nil, 0, fmt.Errorf("invalid request")
	}
	s := lines[0]
	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("invalid request line")
	}
	if !isOnlyUppercase(parts[0]) {
		return nil, 0, fmt.Errorf("invalid request method")
	}
	if strings.HasPrefix(parts[2], "HTTP/") && parts[2][5:] != "1.1" {
		return nil, 0, fmt.Errorf("invalid HTTP version")
	}
	rl := &RequestLine {
		Method: parts[0],
		RequestTarget: parts[1],
		HttpVersion: parts[2][5:],
	}
	return rl, len(s)+2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case ParserStateInitialized:
		requestLine, bytesParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if bytesParsed == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = ParserStateDone
		return bytesParsed, nil
	case ParserStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request {
		State: ParserStateInitialized,
	}
	for req.State != ParserStateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				req.State = ParserStateDone
				break
			}
			return nil, fmt.Errorf("err reading request data from reader")
		}
		readToIndex += numBytesRead
		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		if numBytesParsed > 0 {
			copy(buf, buf[numBytesParsed:readToIndex])
			readToIndex -= numBytesParsed
		}
	}
	return req, nil
}
