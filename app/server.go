package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Printf("Logs from your program will appear here! %v, %v\n", len(os.Args), os.Args)
	fmt.Println(os.Args)

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

		go handleConnection(err, connection)
	}
}

const (
	Host           = "Host"
	UserAgent      = "User-Agent"
	AcceptEncoding = "Accept-Encoding"
	ContentType    = "Content-Type"
	ContentLength  = "Content-Length"
)

var KNOWN_HEADERS = []string{Host, UserAgent, AcceptEncoding, ContentType, ContentLength}

func parseHttpRequest(request []byte) HttpRequest {
	requestData := bytes.Split(request, []byte(LINE_SEPARATOR))

	contentIndex := 0
	for i := 1; i < len(requestData); i++ {
		if len(requestData[i]) == 0 {
			contentIndex = i + 1
			break
		}
	}

	fmt.Printf("Remaining [%d:%d]\n", 1, contentIndex)
	headers := parseHeaders(requestData[1 : contentIndex-1])

	var body []byte
	if length, ok := headers[ContentLength]; ok {
		fmt.Println("Data Length = " + length)
		dataLength, _ := strconv.Atoi(length)
		body = make([]byte, dataLength)
		copy(body, requestData[contentIndex][0:dataLength])
	}

	return HttpRequest{
		StartLine: parseStartLine(requestData[0]),
		Headers:   headers,
		Content:   body,
	}
}

func parseHeaders(requestData [][]byte) map[string]string {
	headers := make(map[string]string)
	for i := 0; i < len(requestData); i++ {
		println("parseHeaders " + string(requestData[i]))
		header := bytes.Split(requestData[i], []byte(": "))
		headers[string(header[0])] = string(header[1])
	}

	return headers
}

func handleConnection(err error, connection net.Conn) {
	// process request

	// content length needs to be exact
	// cannot be 1024
	request := make([]byte, 1024)
	_, err = connection.Read(request)

	httpRequest := parseHttpRequest(request)
	fmt.Printf("httpRequest--Data: ---------->>> %v\n\n", httpRequest)

	switch httpRequest.StartLine.HttpMethod {
	case "GET":
		{
			if httpRequest.StartLine.Path == "/" {
				connection.Write([]byte(("HTTP/1.1 200 OK\r\n\r\n")))
			} else if strings.HasPrefix(httpRequest.StartLine.Path, "/echo/") {
				res := echoResponse(httpRequest.StartLine.Path)
				connection.Write(res)
			} else if strings.HasPrefix(httpRequest.StartLine.Path, "/user-agent") {
				res := userAgent(httpRequest)
				connection.Write(res)
			} else if strings.HasPrefix(httpRequest.StartLine.Path, "/files/") {
				var res = handleFileRead(os.Args[2], httpRequest.StartLine.Path)
				connection.Write(res)
			} else {
				connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
			}
		}
	case "POST":
		{
			if strings.HasPrefix(httpRequest.StartLine.Path, "/files/") {
				filePathAbs := path.Join(os.Args[2], strings.TrimPrefix(httpRequest.StartLine.Path, "/files/"))
				var res = handleFileCreation(filePathAbs, httpRequest)
				connection.Write(res)
			} else {
				connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
			}
		}

	}

	connection.Close()
}

func handleFileCreation(filePathAbs string, httpRequest HttpRequest) []byte {
	fmt.Printf("File Creation Path: [%s] -- data size [%d] \n", filePathAbs, len(httpRequest.Content))
	err := os.WriteFile(filePathAbs, httpRequest.Content, os.ModePerm)
	handleErr(err)

	return []byte("HTTP/1.1 201 CREATED\r\n\r\n")
}

func handleFileRead(directory string, httpPath string) []byte {
	filePath := strings.TrimPrefix(httpPath, "/files/")
	filePathAbs := path.Join(directory, filePath)
	fmt.Println("File Read Path: [%s]", filePathAbs)
	fileContent, err := os.ReadFile(filePathAbs)
	if err != nil {
		fmt.Println("error occurred: ", err.Error())
		return []byte("HTTP/1.1 404 NOT FOUND\r\n\r\n")
	}

	return make200ResponseBytes(fileContent, "application/octet-stream")
}

func userAgent(httpRequest HttpRequest) []byte {
	// skip 0 - it is start line
	if val, ok := httpRequest.Headers[UserAgent]; ok {
		return make200ResponseBytes([]byte(val), "text/plain")
	}
	return []byte{}
}

func echoResponse(path string) []byte {
	content := strings.TrimPrefix(path, "/echo/")
	return make200ResponseBytes([]byte(content), "text/plain")
}

func make200ResponseBytes(content []byte, contentType string) []byte {
	response := make([][]byte, 5)

	response[0] = []byte("HTTP/1.1 200 OK")
	response[1] = []byte(fmt.Sprintf("Content-Type: %s", contentType))
	response[2] = []byte(fmt.Sprintf("Content-Length: %d", len(content)))
	response[3] = []byte(CONTENT_SEPARATOR)
	response[4] = content

	return bytes.Join(response, []byte(LINE_SEPARATOR))
}

func parseStartLine(line []byte) StartLine {
	items := strings.Split(string(line), " ")
	if len(items) != 3 {
		log.Fatal("Expect 'HTTP_METHOD<space>PATH<space>HTTP_VERSION'")
	}
	return StartLine{
		HttpMethod:  items[0],
		Path:        items[1],
		HttpVersion: items[2],
	}
}

type HttpRequest struct {
	StartLine StartLine
	Headers   map[string]string
	Content   []byte
}

type StartLine struct {
	HttpMethod  string
	Path        string
	HttpVersion string
}

//goland:noinspection GoSnakeCaseUsage
const LINE_SEPARATOR = "\r\n"

//goland:noinspection GoSnakeCaseUsage
const CONTENT_SEPARATOR = ""

func handleErr(err error) {
	if err != nil {
		fmt.Println("error occurred: ", err.Error())
		os.Exit(1)
	}
}
