package comet

import (
	"fmt"
	"net/http"
	"strings"
)

type route struct {
	Method      string
	PathPattern string
	Handler     RequestHandler
	PathParts   []string
	ParamNames  []string
}

type CometGroup struct {
	BasePath      string
	StaticRoutes  map[string]RequestHandler
	DynamicRoutes []*route
	Middlewares   []Middleware
}

func Group(basePath string) *CometGroup {
	return &CometGroup{
		BasePath:      basePath,
		StaticRoutes:  make(map[string]RequestHandler),
		DynamicRoutes: make([]*route, 0),
		Middlewares:   make([]Middleware, 0),
	}
}

func (g *CometGroup) Use(middleware Middleware) {
	g.Middlewares = append(g.Middlewares, middleware)
}

func (g *CometGroup) MapGet(path string, handler RequestHandler, middlewares ...Middleware) {
	g.mapRequestHandler(http.MethodGet, path, handler, middlewares...)
}

func (g *CometGroup) MapPost(path string, handler RequestHandler, middlewares ...Middleware) {
	g.mapRequestHandler(http.MethodPost, path, handler, middlewares...)
}

func (g *CometGroup) MapPut(path string, handler RequestHandler, middlewares ...Middleware) {
	g.mapRequestHandler(http.MethodPut, path, handler, middlewares...)
}

func (g *CometGroup) MapPatch(path string, handler RequestHandler, middlewares ...Middleware) {
	g.mapRequestHandler(http.MethodPatch, path, handler, middlewares...)
}

func (g *CometGroup) MapDelete(path string, handler RequestHandler, middlewares ...Middleware) {
	g.mapRequestHandler(http.MethodDelete, path, handler, middlewares...)
}

func (g *CometGroup) mapRequestHandler(method, path string, handler RequestHandler, middlewares ...Middleware) {
	if strings.Contains(path, ":") {
		parts := strings.Split(path, "/")
		params := make([]string, 0)

		for _, part := range parts {
			if strings.HasPrefix(part, ":") {
				params = append(params, part[1:])
			}
		}

		r := &route{
			Method:      method,
			PathPattern: path,
			Handler:     chain(chain(handler, middlewares...), g.Middlewares...),
			PathParts:   parts,
			ParamNames:  params,
		}

		g.DynamicRoutes = append(g.DynamicRoutes, r)
		return
	}

	key := fmt.Sprintf("%s:%s", method, path)
	g.StaticRoutes[key] = chain(chain(handler, middlewares...), g.Middlewares...)
}
