package main

type Header string
type HttpMethod string

type HttpResponse struct {
	ResponseCode int
	Headers      map[Header]string
	Content      []byte
}

type HttpRequest struct {
	StartLine StartLine
	Headers   map[Header]string
	Content   []byte
}

type StartLine struct {
	HttpMethod  HttpMethod
	Path        string
	HttpVersion string
}

type HttpHandler interface {
	Handle(httpRequest HttpRequest) HttpResponse
}

const (
	Host           = Header("Host")
	UserAgent      = Header("User-Agent")
	AcceptEncoding = Header("Accept-Encoding")
	ContentType    = Header("Content-Type")
	ContentLength  = Header("Content-Length")
)
const (
	GET  = HttpMethod("GET")
	POST = HttpMethod("POST")
)
const (
	LineSeparator    = "\r\n"
	ContentSeparator = ""
)

var (
	HTTP_METHOD_PATHS = map[HttpMethod]map[string]HttpHandler{
		GET:  make(map[string]HttpHandler),
		POST: make(map[string]HttpHandler),
	}
	RES_CODE_TO_STATEMENT = map[int]string{
		200: "OK",
		201: "CREATED",
		404: "NOT FOUND",
	}
	KNOWN_HEADERS          = map[Header]struct{}{Host: {}, UserAgent: {}, AcceptEncoding: {}, ContentType: {}, ContentLength: {}}
	SUPPORTED_HTTP_METHODS = map[HttpMethod]struct{}{GET: {}, POST: {}}
)
