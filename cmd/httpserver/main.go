package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func handler(w *response.Writer, req *request.Request) {
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
	err := w.WriteStatusLine(data.StatusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	err = w.WriteHeaders(response.GetHTMLHeaders(len(data.String())))
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	_, err = w.WriteHTMLBody(data)
	if err != nil {
		log.Printf("Error writing body: %v", err)
		return
	}
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
