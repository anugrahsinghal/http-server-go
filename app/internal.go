package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type DefaultHandler struct{}

func (h DefaultHandler) Handle(HttpRequest) HttpResponse {
	return HttpResponse{ResponseCode: 200}
}

func registerHttpHandler(method HttpMethod, pathPrefix string, handler HttpHandler) {
	if _, ok := HTTP_METHOD_PATHS[method][pathPrefix]; ok {
		panic("CANNOT REGISTER ALREADY REGISTERED PATH FOR METHOD " + method)
	}
	HTTP_METHOD_PATHS[method][pathPrefix] = handler
}

func getDispatch(httpRequest HttpRequest) (HttpHandler, error) {
	log.Printf("All Paths:: %v\n", HTTP_METHOD_PATHS[httpRequest.StartLine.HttpMethod])
	if "/" == httpRequest.StartLine.Path { // reg-prefix=/files/ - path=/ - /files/ has /
		return DefaultHandler{}, nil
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
	if _, ok := SUPPORTED_HTTP_METHODS[HttpMethod(items[0])]; !ok {
		log.Fatal(fmt.Sprintf("Only supports %v", SUPPORTED_HTTP_METHODS))
	}
	return StartLine{
		HttpMethod:  HttpMethod(items[0]),
		Path:        items[1],
		HttpVersion: items[2],
	}
}

func parseHeaders(requestData [][]byte) map[Header]string {
	headers := make(map[Header]string)
	for i := 0; i < len(requestData); i++ {
		log.Println("parseHeaders " + string(requestData[i]))
		header := bytes.Split(requestData[i], []byte(": "))
		if _, ok := KNOWN_HEADERS[Header(header[0])]; !ok {
			log.Fatal(fmt.Sprintf("Unknown Header %v", header))
		}
		headers[Header(header[0])] = string(header[1])
	}

	return headers
}

func (res HttpResponse) build() []byte {
	headers := res.formatHeaders()

	var response [][]byte

	response = append(response, []byte(fmt.Sprintf("HTTP/1.1 %d %s", res.ResponseCode, RES_CODE_TO_STATEMENT[res.ResponseCode])))
	response = append(response, headers...)
	response = append(response, []byte(ContentSeparator))
	response = append(response, res.Content)

	join := bytes.Join(response, []byte(LineSeparator))

	fmt.Printf("Response: %s\n", string(join))

	return join
}

func (res HttpResponse) formatHeaders() [][]byte { // assuming res is of type response
	var headers [][]byte
	if _, ok := res.Headers[ContentType]; !ok && len(res.Content) > 0 {
		panic("When Content present then Content-Type header is expected")
	}

	for header, value := range res.Headers {
		if strings.EqualFold(string(header), string(ContentLength)) {
			continue // ignore content length header
		}
		headers = append(headers, []byte(fmt.Sprintf("%s: %s", header, value)))
	}

	headers = append(headers, []byte(fmt.Sprintf("%s: %d", ContentLength, len(res.Content))))
	return headers
}
