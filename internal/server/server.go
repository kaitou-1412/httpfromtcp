package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
	"time"
)

const network = "tcp"

type HandlerError struct {
	StatusCode response.StatusCode
	Message string
}	

type Handler func(w *response.Writer, req *request.Request)

type Server struct{
	listener net.Listener
	closed atomic.Bool
	handler Handler
}	

func Serve(port int, handler Handler) (*Server, error) {
	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", address, err)
	}
	s := &Server{
		listener: listener,
		handler: handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		connection, err := s.listener.Accept()
		if s.closed.Load() {
			return
		}
		if err != nil {
			log.Printf("err accepting connection, %s", err.Error())
			continue
		}
		go s.handle(connection)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("err setting read deadline, %s", err.Error())
		return
	}
	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("err reading request from connection, %s", err.Error())
		hErr := &HandlerError{
			StatusCode: response.BadRequest, 
			Message: "err reading request from connection",
		}
		hErr.Write(w)
		return
	}
	s.handler(w, req)
}

func (h *HandlerError) Write(w *response.Writer) {
	data := response.HTML{
		StatusCode: h.StatusCode,
		H1: response.GetReasonPhrase(h.StatusCode),
		P: h.Message,
	}
	w.WriteStatusLine(data.StatusCode)
	w.WriteHeaders(response.GetHTMLHeaders(len(data.String())))
	w.WriteBody([]byte(data.String()))
}