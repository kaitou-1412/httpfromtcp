package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("err opening file", err)
	}
	linesChannel := getLinesChannel(file)
	for line := range linesChannel {
		fmt.Printf("read: %s\n", line)
	}
}