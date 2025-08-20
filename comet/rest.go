package comet

import (
	"context"
	"net/url"
)

// ControllerBase provides the foundation for all API controllers.
// Offers helper methods for request handling and response generation.
type ControllerBase interface {
	Route() string
	Policies() PoliciesConfig
}

// Request encapsulates the incoming HTTP request with convenient methods
// to access parameters, headers, body content, and other request data.
type Request struct {
	Url           *url.URL
	Method        string
	QueryParams   map[string][]string
	PathParams    map[string]string
	Headers       map[string][]string
	Body          []byte
	UserAgent     string
	RemoteAddress string
	ctx           context.Context
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) WithContext(ctx context.Context) *Request {
	return &Request{
		ctx:         ctx,
		Url:         r.Url,
		Method:      r.Method,
		QueryParams: r.QueryParams,
		PathParams:  r.PathParams,
		Headers:     r.Headers,
		Body:        r.Body,
	}
}

// Response represents the HTTP response to be sent to the client.
// Provides methods to set status codes, headers, and response body content.
type Response struct {
	Status int
	Data   interface{}
}

func Ok[T any](data T) Response {
	return Response{
		Status: 200,
		Data:   data,
	}
}

func Error[T any](data T) Response {
	return Response{
		Status: 500,
		Data:   data,
	}
}

func NotFound() Response {
	return Response{
		Status: 404,
		Data:   "resource not found",
	}
}

func BadRequest[T any](data T) Response {
	return Response{
		Status: 400,
		Data:   data,
	}
}

func Unauthorized() Response {
	return Response{
		Status: 401,
		Data:   "Unauthorized",
	}
}

// RequestHandler is a function type that processes HTTP requests and generates responses.
// The fundamental building block for defining API endpoints and handlers.
type RequestHandler func(*Request) Response
