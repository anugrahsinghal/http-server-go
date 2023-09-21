package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	Host             = "Host"
	UserAgent        = "User-Agent"
	AcceptEncoding   = "Accept-Encoding"
	ContentType      = "Content-Type"
	ContentLength    = "Content-Length"
	LineSeparator    = "\r\n"
	ContentSeparator = ""
)

var (
	HTTP_METHOD_PATHS = make(map[string]map[string]func(httpRequest HttpRequest) []byte)
	KNOWN_HEADERS     = []string{Host, UserAgent, AcceptEncoding, ContentType, ContentLength}
)

func register(method string, pathPrefix string, function func(httpRequest HttpRequest) []byte) {
	if HTTP_METHOD_PATHS[method] == nil {
		HTTP_METHOD_PATHS[method] = make(map[string]func(httpRequest HttpRequest) []byte)
	}
	if _, ok := HTTP_METHOD_PATHS[method][pathPrefix]; ok {
		panic("CANNOT REGISTER ALREADY REGISTERED PATH FOR METHOD " + method)
	}
	HTTP_METHOD_PATHS[method][pathPrefix] = function
}

func getDispatch(httpRequest HttpRequest) (func(httpRequest HttpRequest) []byte, error) {
	log.Printf("All Paths:: %v\n", HTTP_METHOD_PATHS[httpRequest.StartLine.HttpMethod])
	if "/" == httpRequest.StartLine.Path { // reg-prefix=/files/ - path=/ - /files/ has /
		return func(httpRequest HttpRequest) []byte {
			return []byte(("HTTP/1.1 200 OK\r\n\r\n"))
		}, nil
	}
	for registeredPrefix, dispatch := range HTTP_METHOD_PATHS[httpRequest.StartLine.HttpMethod] {
		log.Printf("Compare Registered Path %s to %s \n", registeredPrefix, httpRequest.StartLine.Path)
		// not exactly correct - because we want best match
		if strings.HasPrefix(httpRequest.StartLine.Path, registeredPrefix) { // reg-prefix=/files/ - path=/ - /files/ has /
			log.Println("Dispatch to : " + registeredPrefix)
			return dispatch, nil
		}
	}
	return nil, errors.New("Path Not Found")
}

func parseHttpRequest(request []byte) HttpRequest {
	requestData := bytes.Split(request, []byte(LineSeparator))

	contentIndex := 0
	for i := 1; i < len(requestData); i++ {
		if len(requestData[i]) == 0 {
			contentIndex = i + 1
			break
		}
	}

	log.Printf("Remaining [%d:%d]\n", 1, contentIndex)
	headers := parseHeaders(requestData[1 : contentIndex-1])

	var body []byte
	if length, ok := headers[ContentLength]; ok {
		log.Println("Data Length = " + length)
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
		log.Println("parseHeaders " + string(requestData[i]))
		header := bytes.Split(requestData[i], []byte(": "))
		headers[string(header[0])] = string(header[1])
	}

	return headers
}

func makeSuccessResponse(content []byte, contentType string) []byte {
	response := make([][]byte, 5)

	response[0] = []byte("HTTP/1.1 200 OK")
	response[1] = []byte(fmt.Sprintf("Content-Type: %s", contentType))
	response[2] = []byte(fmt.Sprintf("Content-Length: %d", len(content)))
	response[3] = []byte(ContentSeparator)
	response[4] = content

	return bytes.Join(response, []byte(LineSeparator))
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
