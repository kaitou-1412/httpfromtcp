package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
	address, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("err resolving UDP address", err)
	}
	connection, err:= net.DialUDP("udp", nil, address)
	if err != nil {
		log.Fatal("err listening to port 42069", err)
	}
	defer connection.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("err reading", err)
		}
		connection.Write([]byte(input))
	}
}