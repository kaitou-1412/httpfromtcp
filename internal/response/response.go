package response

import (
	"fmt"
	"httpfromtcp/internal/header"
	"io"
	"strconv"
	"strings"
)

const crlf = "\r\n"

type StatusCode int

const (
	Ok StatusCode = 200
	BadRequest StatusCode = 400
	InternalServerError StatusCode = 500
)

type HTML struct {
	StatusCode StatusCode
	H1 string
	P string
}

type WriterState string

const (
	WriterStateInitialized WriterState = "init"
	WriterStateStatusLine WriterState = "statusline"
	WriterStateHeader WriterState = "header"
	WriterStateBody WriterState = "body"
)

type Writer struct{
	writer io.Writer
	state WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state: WriterStateInitialized,
	}
}

func GetReasonPhrase(statusCode StatusCode) string {
	var reasonPhrase string
	switch statusCode {
	case Ok:
		reasonPhrase = "OK"
	case BadRequest:
		reasonPhrase = "Bad Request"
	case InternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}
	return reasonPhrase
}

func GetDefaultHeaders(contentLen int) header.Headers {
	return header.Headers {
		"Content-Length": strconv.Itoa(contentLen),
		"Connection": "close",
		"Content-Type": "text/plain",
	}
}

func GetHTMLHeaders(contentLen int) header.Headers {
	headers := GetDefaultHeaders(contentLen)
	headers["Content-Type"] = "text/html"
	return headers
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != WriterStateInitialized {
		return fmt.Errorf("invalid state: expected %s, got %s", WriterStateInitialized, w.state)
	}
	w.state = WriterStateStatusLine
	reasonPhrase := GetReasonPhrase(statusCode)
	_, err := fmt.Fprintf(w.writer, "HTTP/1.1 %d %s%s", statusCode, reasonPhrase, crlf)
	return err
}

func (w *Writer) WriteHeaders(headers header.Headers) error {
	if w.state != WriterStateStatusLine {
		return fmt.Errorf("invalid state: expected %s, got %s", WriterStateStatusLine, w.state)
	}
	w.state = WriterStateHeader
	var b strings.Builder
	for key, value := range(headers) {
		fmt.Fprintf(&b, "%s: %s%s", key, value, crlf)
	}
	b.WriteString(crlf)
	_, err := io.WriteString(w.writer, b.String())
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != WriterStateHeader && w.state != WriterStateBody {
		return 0, fmt.Errorf("invalid state: expected %s, got %s", WriterStateHeader, w.state)
	}
	w.state = WriterStateBody
	n, err := w.writer.Write(p)
	return n, err
}

func (data HTML) String() string {
	return fmt.Sprintf(
		"<html><head><title>%d %s</title></head><body><h1>%s</h1><p>%s</p></body></html>",
		data.StatusCode,
		GetReasonPhrase(data.StatusCode),
		data.H1,
		data.P,
	)
}

func (w *Writer) WriteHTMLBody(data HTML) (int, error) {
	p := []byte(data.String())
	n, err := w.WriteBody(p)
	return n, err
}
