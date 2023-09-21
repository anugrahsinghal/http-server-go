package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		request := make([]byte, 1024)
		_, err = connection.Read(request)
		requestData := string(request)
		fmt.Printf("Data: \n%s", requestData)
		splitData := strings.Split(requestData, "\r\n")
		for i, datum := range splitData {
			fmt.Printf("i=%v -- Data: %v----\n", i, datum)
		}
		startLine := parseStartLine(splitData[0])
		if startLine.Path == "/" {
			connection.Write([]byte(("HTTP/1.1 200 OK\r\n\r\n")))
		} else {
			connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
		}

		connection.Close()
	}
}

func parseStartLine(line string) StartLine {
	items := strings.Split(line, " ")
	if len(items) != 3 {
		log.Fatal("Expect 'HTTP_METHOD<space>PATH<space>HTTP_VERSION'")
	}
	return StartLine{
		HttpMethod:  items[0],
		Path:        items[1],
		HttpVersion: items[2],
	}
}

type StartLine struct {
	HttpMethod  string
	Path        string
	HttpVersion string
}

func handleErr(err error) {
	if err != nil {
		fmt.Println("error occurred: ", err.Error())
		os.Exit(1)
	}
}
