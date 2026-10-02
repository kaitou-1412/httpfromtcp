package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)
		part := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err == io.EOF {
				break
			}
			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				part += string(data[:i])
				data = data[i+1:]
				out <- part
				part = ""
			}
			part += string(data)
		}
		if len(part) != 0 {
			out <- part
		}
	}()
	return out
}

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
		linesChannel := getLinesChannel(connection)
		for line := range linesChannel {
			fmt.Printf("read: %s\n", line)
		}
		connection.Close()
	}
}