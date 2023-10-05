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
	registerHttpHandler(GET, "/echo/", EchoHandler{})
	registerHttpHandler(GET, "/user-agent", UserAgentHandler{})
	registerHttpHandler(GET, "/files/", FileReadHandler{})
	registerHttpHandler(POST, "/files/", FileCreationHandler{})
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	request := make([]byte, 1024)
	_, err := conn.Read(request)
	if err != nil {
		log.Println("Failed to read request: ", err.Error())
		return
	}

	httpRequest := parseHttpRequest(request)
	log.Printf("httpRequest--Data: ---------->>> %v\n\n", httpRequest)

	var response HttpResponse

	dispatch, err := getDispatch(httpRequest)
	if err != nil {
		log.Println("Dispatch error occurred: ", err.Error())
		response = HttpResponse{StatusCode: 404}
	} else {
		response = dispatch.Handle(httpRequest)
	}

	conn.Write(response.build())
}

type FileCreationHandler struct{}

func (h FileCreationHandler) Handle(httpRequest HttpRequest) HttpResponse {
	filePathAbs := path.Join(os.Args[2], strings.TrimPrefix(httpRequest.StartLine.Path, "/files/"))
	err := os.WriteFile(filePathAbs, httpRequest.Content, os.ModePerm)
	handleErr(err)

	return HttpResponse{StatusCode: 201}
}

type FileReadHandler struct{}

func (h FileReadHandler) Handle(httpRequest HttpRequest) HttpResponse {
	filePath := strings.TrimPrefix(httpRequest.StartLine.Path, "/files/")
	filePathAbs := path.Join(os.Args[2], filePath)
	fileContent, err := os.ReadFile(filePathAbs)
	if err != nil {
		log.Println("error occurred: ", err.Error())
		return HttpResponse{StatusCode: 404}
	}

	return HttpResponse{
		StatusCode: 200,
		Headers:    map[Header]string{ContentType: "application/octet-stream"},
		Content:    fileContent,
	}
}

type UserAgentHandler struct{}

func (h UserAgentHandler) Handle(httpRequest HttpRequest) HttpResponse {
	if _, ok := httpRequest.Headers[UserAgent]; !ok {
		return HttpResponse{}
	}

	return HttpResponse{
		StatusCode: 200,
		Headers:    map[Header]string{ContentType: "text/plain"},
		Content:    []byte(httpRequest.Headers[UserAgent]),
	}
}

type EchoHandler struct{}

func (h EchoHandler) Handle(httpRequest HttpRequest) HttpResponse {
	content := strings.TrimPrefix(httpRequest.StartLine.Path, "/echo/")
	return HttpResponse{
		StatusCode: 200,
		Headers:    map[Header]string{ContentType: "text/plain"},
		Content:    []byte(content),
	}
}

func handleErr(err error) {
	if err != nil {
		log.Println("error occurred: ", err.Error())
		os.Exit(1)
	}
}
