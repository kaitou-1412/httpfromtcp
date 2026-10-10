package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/header"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

func writeHTML(w *response.Writer, data response.HTML) {
	if err := w.WriteStatusLine(data.StatusCode); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	if err := w.WriteHeaders(response.GetHTMLHeaders(len(data.String()))); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	if _, err := w.WriteHTMLBody(data); err != nil {
		log.Printf("Error writing body: %v", err)
		return
	}
}

func proxyHandler(w *response.Writer, req *request.Request) bool {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		return false
	}
	var body []byte
	data := response.HTML{}
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	proxyURL := "https://httpbingo.org/" + path
	resp, err := http.Get(proxyURL)
	if err != nil {
		data.StatusCode = response.InternalServerError
		data.H1 = response.GetReasonPhrase(data.StatusCode)
		data.P = "Hey, is httpbingo.org down?"
		writeHTML(w, data)
		return true
	}
	defer resp.Body.Close()
	data.StatusCode = response.Ok
	if err := w.WriteStatusLine(data.StatusCode); err != nil {
		log.Printf("Error writing status line: %v", err)
		return true
	}
	h := response.GetChunkedHeaders()
	h["Trailer"] = "X-Content-SHA256, X-Content-Length"
	if err := w.WriteHeaders(h); err != nil {
		log.Printf("Error writing headers: %v", err)
		return true
	}
	chunk := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(chunk)
		if n > 0 {
			log.Printf("read %d bytes", n)
			if _, err := w.WriteChunkedBody(chunk[:n]); err != nil {
				log.Printf("Error writing chunked body: %v", err)
				return true
			}
			body = append(body, chunk[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			break
		}
	}
	if _, err := w.WriteChunkedBodyDone(); err != nil {
		log.Printf("Error writing last chunk: %v", err)
		return true
	}
	checksum := sha256.Sum256(body)
	trailers := header.Headers{
		"X-Content-SHA256": fmt.Sprintf("%x", checksum),
		"X-Content-Length": strconv.Itoa(len(body)),
	}
	if err := w.WriteTrailers(trailers); err != nil {
		log.Printf("Error writing trailers: %v", err)
		return true
	}
	return true
}

func videoHandler(w *response.Writer, req *request.Request) bool {
	if req.RequestLine.Method != "GET" || req.RequestLine.RequestTarget != "/video" {
		return false
	}
	file, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		writeHTML(w, response.HTML{
			StatusCode: response.InternalServerError,
			H1: response.GetReasonPhrase(response.InternalServerError),
			P: "Error reading video file",
		})
		return true
	}
	if err := w.WriteStatusLine(response.Ok); err != nil {
		log.Printf("Error writing status line: %v", err)
		return true
	}
	h := response.GetDefaultHeaders(len(file))
	h["Content-Type"] = "video/mp4" 
	if err := w.WriteHeaders(h); err != nil {
		log.Printf("Error writing headers: %v", err)
		return true
	}
	if _, err := w.WriteBody(file); err != nil {
		log.Printf("Error writing video body: %v", err)
		return true
	}
	return true
}

func handler(w *response.Writer, req *request.Request) {
	if proxyHandler(w, req) {
		return
	} else if videoHandler(w, req) {
		return
	}
	data := response.HTML{}
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		data.StatusCode = response.BadRequest
		data.H1 = response.GetReasonPhrase(data.StatusCode)
		data.P = "Your request honestly kinda sucked."
	case "/myproblem":
		data.StatusCode = response.InternalServerError
		data.H1 = response.GetReasonPhrase(data.StatusCode)
		data.P = "Okay, you know what? This one is on me."
	default:
		data.StatusCode = response.Ok
		data.H1 = "Success!"
		data.P = "Your request was an absolute banger."
	}
	writeHTML(w, data)
}

func main() {
    srv, err := server.Serve(port, handler)
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
    defer srv.Close()
    log.Println("Server started on port", port)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    log.Println("Server gracefully stopped")
}
