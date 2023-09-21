package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
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

func handleConnection(err error, connection net.Conn) {
	// process request
	request := make([]byte, 1024)
	_, err = connection.Read(request)
	requestData := string(request)
	fmt.Printf("Data: \n%s", requestData)
	httpData := strings.Split(requestData, "\r\n")
	for i, datum := range httpData {
		fmt.Printf("i=%v -- Data: %v----\n", i, datum)
	}

	startLine := parseStartLine(httpData[0])

	if startLine.Path == "/" {
		connection.Write([]byte(("HTTP/1.1 200 OK\r\n\r\n")))
	} else if strings.HasPrefix(startLine.Path, "/echo/") {
		res := echoResponse(startLine.Path)
		connection.Write([]byte((res)))
	} else if strings.HasPrefix(startLine.Path, "/user-agent") {
		res := userAgent(httpData)
		connection.Write([]byte((res)))
	} else if strings.HasPrefix(startLine.Path, "/files/") {
		var res = handleFiles(os.Args[2], startLine.Path)
		connection.Write([]byte((res)))
	} else {
		connection.Write([]byte(("HTTP/1.1 404 NOT FOUND\r\n\r\n")))
	}

	connection.Close()
}

func handleFiles(directory string, httpPath string) string {
	filePath := strings.TrimPrefix(httpPath, "/files/")
	filePathAbs := path.Join(directory, filePath)
	fileContent, err := os.ReadFile(filePathAbs)
	//file, err := os.OpenFile(filePathAbs, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println("error occurred: ", err.Error())
		return "HTTP/1.1 404 NOT FOUND\r\n\r\n"
	}

	//fileBuffer := bufio.NewReader(file)
	//fileSize := fileBuffer.Size()
	//fileContent := make([]byte, fileSize)
	//_, err = io.ReadFull(fileBuffer, fileContent)
	//handleErr(err)

	return make200Response(string(fileContent), "application/octet-stream")
}

func userAgent(httpData []string) string {
	// skip 0 - it is start line
	for i := 1; i < len(httpData); i++ {
		if ok := strings.HasPrefix(httpData[i], "User-Agent:"); ok {
			return make200Response(strings.TrimPrefix(httpData[i], "User-Agent: "), "text/plain")
		}
	}
	return ""
}

func echoResponse(path string) string {
	content := strings.TrimPrefix(path, "/echo/")
	return make200Response(content, "text/plain")
}

func make200Response(content string, contentType string) string {
	response := make([]string, 5)

	response[0] = "HTTP/1.1 200 OK"
	response[1] = fmt.Sprintf("Content-Type: %s", contentType)
	response[2] = fmt.Sprintf("Content-Length: %d", len(content))
	response[3] = CONTENT_SEPARATOR
	response[4] = content

	return strings.Join(response, LINE_SEPARATOR)
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
