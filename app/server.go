package main

import (
	"bytes"
	"errors"
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

var (
	HTTP_METHOD_PATHS = make(map[string]map[string]func(httpRequest HttpRequest) []byte)
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

	registerPaths()

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(err, connection)
	}
}

func registerPaths() {
	register("GET", "/echo/", echoResponse)
	register("GET", "/user-agent", userAgent)
	register("GET", "/files/", handleFileRead)
	register("POST", "/files/", handleFileCreation)
}

func register(method string, path string, function func(httpRequest HttpRequest) []byte) {
	if HTTP_METHOD_PATHS[method] == nil {
		HTTP_METHOD_PATHS[method] = make(map[string]func(httpRequest HttpRequest) []byte)
	}
	HTTP_METHOD_PATHS[method][path] = function
}

func getDispatch(httpRequest HttpRequest) (func(httpRequest HttpRequest) []byte, error) {
	fmt.Printf("All Paths:: %v\n", HTTP_METHOD_PATHS[httpRequest.StartLine.HttpMethod])
	for registeredPrefix, dispatch := range HTTP_METHOD_PATHS[httpRequest.StartLine.HttpMethod] {
		fmt.Printf("Compare Registered Path %s to %s \n", registeredPrefix, httpRequest.StartLine.Path)
		// not exactly correct - because we want best match
		if strings.HasPrefix(httpRequest.StartLine.Path, registeredPrefix) { // reg-prefix=/files/ - path=/ - /files/ has /
			fmt.Println("Dispatch to : " + registeredPrefix)
			return dispatch, nil
		}
	}
	return nil, errors.New("Path Not Found")
}

func handleConnection(err error, connection net.Conn) {
	request := make([]byte, 1024)
	_, err = connection.Read(request)
	defer connection.Close()

	httpRequest := parseHttpRequest(request)
	fmt.Printf("httpRequest--Data: ---------->>> %v\n\n", httpRequest)

	if "/" == httpRequest.StartLine.Path { // reg-prefix=/files/ - path=/ - /files/ has /
		connection.Write([]byte(("HTTP/1.1 200 OK\r\n\r\n")))
		return
	}

	dispatch, err := getDispatch(httpRequest)
	if err != nil {
		fmt.Println("Dispatch error occurred: ", err.Error())
		connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
		return
	}

	response := dispatch(httpRequest)

	connection.Write(response)

	//switch httpRequest.StartLine.HttpMethod {
	//case "GET":
	//	{
	//		if httpRequest.StartLine.Path == "/" {
	//			connection.Write([]byte(("HTTP/1.1 200 OK\r\n\r\n")))
	//		} else if strings.HasPrefix(httpRequest.StartLine.Path, "/echo/") {
	//			res := echoResponse(httpRequest.StartLine.Path)
	//			connection.Write(res)
	//		} else if strings.HasPrefix(httpRequest.StartLine.Path, "/user-agent") {
	//			res := userAgent(httpRequest)
	//			connection.Write(res)
	//		} else if strings.HasPrefix(httpRequest.StartLine.Path, "/files/") {
	//			var res = handleFileRead(os.Args[2], httpRequest.StartLine.Path)
	//			connection.Write(res)
	//		} else {
	//			connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
	//		}
	//	}
	//case "POST":
	//	{
	//		if strings.HasPrefix(httpRequest.StartLine.Path, "/files/") {
	//			filePathAbs := path.Join(os.Args[2], strings.TrimPrefix(httpRequest.StartLine.Path, "/files/"))
	//			var res = handleFileCreation(filePathAbs, httpRequest)
	//			connection.Write(res)
	//		} else {
	//			connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
	//		}
	//	}
	//
	//}

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

func echoResponse(httpRequest HttpRequest) []byte {
	content := strings.TrimPrefix(httpRequest.StartLine.Path, "/echo/")
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

func parseHeaders(requestData [][]byte) map[string]string {
	headers := make(map[string]string)
	for i := 0; i < len(requestData); i++ {
		println("parseHeaders " + string(requestData[i]))
		header := bytes.Split(requestData[i], []byte(": "))
		headers[string(header[0])] = string(header[1])
	}

	return headers
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
