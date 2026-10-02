package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatal("err listening to port 42069", err)
	}
	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal("err accepting connection", err)
		}
		req, err := request.RequestFromReader(connection)
		if err != nil {
			log.Fatal("err reading request from connection", err)
		}
		fmt.Println("Request Line")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		connection.Close()
	}
}