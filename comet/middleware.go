package comet

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// Middleware is a function that intercepts and processes HTTP requests
// before they reach the main handler, enabling cross-cutting concerns.
type Middleware = func(next RequestHandler) RequestHandler

var RequestLogging Middleware = func(next RequestHandler) RequestHandler {
	return func(req *Request) Response {
		logger := FromContext(req.Context())

		logger.Debug("request received", "method", req.Method, "path", req.Url.Path)

		return next(req)
	}
}

type panicLog struct {
	Timestamp     time.Time   `json:"timestamp"`
	URL           string      `json:"url"`
	Method        string      `json:"method"`
	Error         interface{} `json:"error"`
	StackTrace    string      `json:"stack_trace"`
	UserAgent     string      `json:"user_agent"`
	RemoteAddress string      `json:"remote_addr"`
}

var Recover Middleware = func(next RequestHandler) RequestHandler {
	return func(req *Request) Response {
		defer func(req *Request) {
			if err := recover(); err != nil {
				panicLog := panicLog{
					Timestamp:     time.Now(),
					URL:           req.Url.String(),
					Method:        req.Method,
					Error:         err,
					StackTrace:    string(debug.Stack()),
					UserAgent:     req.UserAgent,
					RemoteAddress: req.RemoteAddress,
				}

				logger := FromContext(req.Context())

				logger.Error("panic error received from request", "info", panicLog)
			}
		}(req)
		return next(req)
	}
}

var HTTPAdapter = func(next RequestHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error parsing body", 500)
			return
		}

		request := Request{
			Url:           r.URL,
			Method:        r.Method,
			QueryParams:   r.URL.Query(),
			PathParams:    make(map[string]string),
			Headers:       r.Header,
			Body:          bytes,
			UserAgent:     r.UserAgent(),
			RemoteAddress: r.RemoteAddr,
		}

		ctx := context.WithValue(r.Context(), TRACE_ID, uuid.New().String())

		response := next(request.WithContext(ctx))

		responseBytes, err := json.Marshal(response.Data)
		if err != nil {
			http.Error(w, "error deserializing response", 500)
			return
		}

		w.WriteHeader(response.Status)
		w.Write(responseBytes)
	})
}

var RequestID Middleware = func(next RequestHandler) RequestHandler {
	return func(req *Request) Response {
		ctx := context.WithValue(req.Context(), "X-Request-Id", uuid.New().String())
		return next(req.WithContext(ctx))
	}
}
