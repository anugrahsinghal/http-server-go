package main

import (
	"log"
	"net"
	"os"
	"path"
	"strings"
)

func main() {
	log.Printf("Logs from your program will appear here! %v, %v\n", len(os.Args), os.Args)

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	registerPaths()

	for {
		connection, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(connection)
	}
}

func registerPaths() {
	registerHttpDispatch("GET", "/echo/", echoResponse)
	registerHttpDispatch("GET", "/user-agent", userAgent)
	registerHttpDispatch("GET", "/files/", handleFileRead)
	registerHttpDispatch("POST", "/files/", handleFileCreation)
}

func handleConnection(conn net.Conn) {
	request := make([]byte, 1024)
	_, err := conn.Read(request)
	defer conn.Close()

	httpRequest := parseHttpRequest(request)
	log.Printf("httpRequest--Data: ---------->>> %v\n\n", httpRequest)

	dispatch, err := getDispatch(httpRequest)
	if err != nil {
		log.Println("Dispatch error occurred: ", err.Error())
		conn.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
		return
	}

	response := dispatch(httpRequest)

	conn.Write(response)
}

func handleFileCreation(httpRequest HttpRequest) []byte {
	filePathAbs := path.Join(os.Args[2], strings.TrimPrefix(httpRequest.StartLine.Path, "/files/"))
	err := os.WriteFile(filePathAbs, httpRequest.Content, os.ModePerm)
	handleErr(err)

	return []byte("HTTP/1.1 201 CREATED\r\n\r\n")
}

func handleFileRead(httpRequest HttpRequest) []byte {
	filePath := strings.TrimPrefix(httpRequest.StartLine.Path, "/files/")
	filePathAbs := path.Join(os.Args[2], filePath)
	fileContent, err := os.ReadFile(filePathAbs)
	if err != nil {
		log.Println("error occurred: ", err.Error())
		return []byte("HTTP/1.1 404 NOT FOUND\r\n\r\n")
	}

	return HttpResponse{
		ResponseCode: 200,
		Headers:      map[Header]string{ContentType: "application/octet-stream"},
		Content:      fileContent,
	}.build()
}

func userAgent(httpRequest HttpRequest) []byte {
	if _, ok := httpRequest.Headers[UserAgent]; !ok {
		return []byte{}
	}

	return HttpResponse{
		ResponseCode: 200,
		Headers:      map[Header]string{ContentType: "text/plain"},
		Content:      []byte(httpRequest.Headers[UserAgent]),
	}.build()
}

func echoResponse(httpRequest HttpRequest) []byte {
	content := strings.TrimPrefix(httpRequest.StartLine.Path, "/echo/")
	return HttpResponse{
		ResponseCode: 200,
		Headers:      map[Header]string{ContentType: "text/plain"},
		Content:      []byte(content),
	}.build()
}

func handleErr(err error) {
	if err != nil {
		log.Println("error occurred: ", err.Error())
		os.Exit(1)
	}
}
