package comet

import (
	"fmt"
	"strings"
)

type router struct {
	groups map[string]*CometGroup
}

func (r *router) Handle(req *Request) Response {
	path := req.Url.Path

	var group *CometGroup = nil
	for key, val := range r.groups {
		if strings.HasPrefix(path, key) {
			group = val
			break
		}
	}

	if group == nil {
		return NotFound()
	}

	path = strings.Replace(path, group.BasePath, "", 1)

	var handler RequestHandler
	handler, ok := group.StaticRoutes[fmt.Sprintf("%s:%s", req.Method, path)]
	if !ok {
		for _, route := range group.DynamicRoutes {
			if route.Method != req.Method {
				continue
			}

			if params, ok := r.matchPath(*route, path); ok {
				req.PathParams = params
				handler = route.Handler
			}
		}
	}

	if handler == nil {
		return NotFound()
	}

	return handler(req)
}

func newRouter() *router {
	return &router{
		groups: map[string]*CometGroup{
			"default": {
				BasePath:      "",
				StaticRoutes:  make(map[string]RequestHandler),
				DynamicRoutes: make([]*route, 0),
				Middlewares:   make([]Middleware, 0),
			},
		},
	}
}

func (r *router) matchPath(route route, path string) (map[string]string, bool) {
	pathParts := strings.Split(path, "/")

	if len(pathParts) != len(route.PathParts) {
		return nil, false
	}

	params := make(map[string]string)

	for i, part := range route.PathParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			params[paramName] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, false
		}
	}

	return params, true
}

func chain(handler RequestHandler, middlewares ...Middleware) RequestHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
