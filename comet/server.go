package comet

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

type Router struct {
	Address     string
	server      *http.ServeMux
	router      *router
	middlewares []Middleware
}

func NewDefaultRouter() *Router {
	return &Router{
		Address:     ":5051",
		server:      http.NewServeMux(),
		router:      newRouter(),
		middlewares: make([]Middleware, 0),
	}
}

func (r *Router) MapGet(path string, handler RequestHandler, middlewares ...Middleware) {
	r.router.groups["default"].mapRequestHandler(http.MethodGet, path, handler, middlewares...)
}

func (r *Router) MapPost(path string, handler RequestHandler, middlewares ...Middleware) {
	r.router.groups["default"].mapRequestHandler(http.MethodPost, path, handler, middlewares...)
}

func (r *Router) MapPut(path string, handler RequestHandler, middlewares ...Middleware) {
	r.router.groups["default"].mapRequestHandler(http.MethodPut, path, handler, middlewares...)
}

func (r *Router) MapPatch(path string, handler RequestHandler, middlewares ...Middleware) {
	r.router.groups["default"].mapRequestHandler(http.MethodPatch, path, handler, middlewares...)
}

func (r *Router) MapDelete(path string, handler RequestHandler, middlewares ...Middleware) {
	r.router.groups["default"].mapRequestHandler(http.MethodDelete, path, handler, middlewares...)
}

func (r *Router) MapGroup(group *CometGroup) {
	r.router.groups[group.BasePath] = group
}

func (r *Router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) MapController(controller ControllerBase, middlewares ...Middleware) {
	controllerType := reflect.TypeOf(controller)

	basePath := controller.Route()
	if basePath == "" {
		basePath = getControllerBaseRoute(controllerType.Name())
	}

	dynamicRoutes := make([]*route, 0)
	staticRoutes := make(map[string]RequestHandler)

	policies := controller.Policies()
	globalPolicies := policies["*"]

	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		if !isRequestMethod(method) {
			continue
		}

		invariantName := strings.ToUpper(method.Name)
		methodMap := []requestMethod{get, post, delete, patch, put, list}

		for _, prefix := range methodMap {
			if !strings.HasPrefix(invariantName, prefix.string()) {
				continue
			}

			path := getMethodPath(basePath, method.Name)
			httpMethod := prefix.method()
			handler := func(r *Request) Response {
				response := method.Func.Call([]reflect.Value{
					reflect.ValueOf(controller),
					reflect.ValueOf(r),
				})

				return response[0].Interface().(Response)
			}

			methodPolicies := policies[method.Name]
			config := make([]Policy, 0)
			if globalPolicies != nil {
				config = append(config, globalPolicies...)
			}

			if methodPolicies != nil {
				config = append(config, methodPolicies...)
			}

			policyMap := make(map[interface{}]AuthorizerFunction)
			for _, val := range config {
				policyMap[val.Value] = val.Validation
			}

			if !strings.Contains(path, ":") {
				key := fmt.Sprintf("%s:%s", httpMethod, strings.Replace(path, basePath, "", 1))
				handler = chainAuthorizations(handler, policyMap)
				staticRoutes[key] = chain(handler, middlewares...)
				break
			}

			parts := strings.Split(path, "/")
			params := make([]string, 0)

			for _, part := range parts {
				if strings.HasPrefix(part, ":") {
					params = append(params, part[1:])
				}
			}

			handler = chainAuthorizations(handler, policyMap)

			dynamicRoutes = append(dynamicRoutes, &route{
				Method:      httpMethod,
				PathPattern: path,
				Handler:     chain(handler, middlewares...),
				PathParts:   parts,
				ParamNames:  params,
			})
		}

	}

	r.router.groups[basePath] = &CometGroup{
		BasePath:      basePath,
		StaticRoutes:  staticRoutes,
		DynamicRoutes: dynamicRoutes,
		Middlewares:   make([]Middleware, 0),
	}
}

func (r *Router) Run() error {
	fmt.Printf("Starting server in %s...\n", r.Address)
	fmt.Println("Routes...")
	for _, group := range r.router.groups {
		for _, route := range group.DynamicRoutes {
			fmt.Printf("[%s]: %s%s\n", route.Method, group.BasePath, route.PathPattern)
		}

		for key := range group.StaticRoutes {
			method := strings.Split(key, ":")[0]
			path := strings.Replace(key, method+":", "", 1)
			fmt.Printf("[%s]: %s%s\n", method, group.BasePath, path)
		}
	}

	middlewares := chain(r.router.Handle, r.middlewares...)
	r.server.Handle("/", httpAdapter(middlewares))

	return http.ListenAndServe(r.Address, r.server)
}

var httpAdapter = func(next RequestHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error parsing body", 500)
			return
		}

		request := &Request{
			Url:           r.URL,
			Method:        r.Method,
			QueryParams:   r.URL.Query(),
			PathParams:    make(map[string]string),
			Headers:       r.Header,
			Body:          bytes,
			UserAgent:     r.UserAgent(),
			RemoteAddress: r.RemoteAddr,
		}

		response := next(request)

		responseBytes, err := json.Marshal(response.Data)
		if err != nil {
			http.Error(w, "error deserializing response", 500)
			return
		}

		w.WriteHeader(response.Status)
		w.Write(responseBytes)
	})
}

func getControllerBaseRoute(baseName string) string {
	name := strings.ReplaceAll(baseName, "Controller", "")
	var result strings.Builder
	for i, char := range name {
		if unicode.IsUpper(char) {
			if i > 0 && !unicode.IsUpper(rune(name[i-1])) {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(char))
		} else {
			result.WriteRune(char)
		}
	}
	name = result.String()
	return "/" + name
}

func isRequestMethod(method reflect.Method) bool {
	if method.Name[0] < 'A' || method.Name[0] > 'Z' {
		return false
	}

	validPrefixes := []requestMethod{get, post, put, patch, delete, list}

	methodName := strings.ToUpper(method.Name)
	matchPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(methodName, prefix.string()) {
			matchPrefix = true
			break
		}
	}

	if !matchPrefix {
		return false
	}

	if method.Type.NumIn() != 2 {
		return false
	}

	reqType := method.Type.In(1)
	if reqType.Kind() != reflect.Ptr {
		return false
	}

	reqType = reqType.Elem()

	if method.Type.NumOut() != 1 {
		return false
	}

	respType := method.Type.Out(0)
	if respType.Kind() == reflect.Ptr {
		respType = respType.Elem()
	}

	if respType != reflect.TypeOf(Response{}) {
		return false
	}

	return reqType.Name() == "Request" && strings.Contains(reqType.PkgPath(), "github.com/ramoncl001/go-comet/rest")
}

func getMethodPath(basePath, methodName string) string {
	httpMethods := []string{"Get", "Post", "Put", "Delete", "Patch", "List"}
	for _, prefix := range httpMethods {
		if strings.HasPrefix(methodName, prefix) {
			methodName = strings.TrimPrefix(methodName, prefix)
			break
		}
	}

	var result strings.Builder
	for i, char := range methodName {
		if unicode.IsUpper(char) {
			if i > 0 && !unicode.IsUpper(rune(methodName[i-1])) {
				result.WriteRune('/')
			}
			result.WriteRune(unicode.ToLower(char))
		} else {
			result.WriteRune(char)
		}
	}
	route := result.String()

	re := regexp.MustCompile(`(?:^|/)(by|for|of|with)/([^/]+)`)
	route = re.ReplaceAllStringFunc(route, func(match string) string {
		parts := strings.Split(match, "/")
		paramName := parts[len(parts)-1] // El último segmento es el parámetro

		if strings.HasPrefix(match, "/") {
			return "/:" + paramName
		}
		return ":" + paramName
	})

	baseName := strings.ReplaceAll(basePath, "/", "")
	route = strings.ReplaceAll(route, baseName+"/", "")
	route = strings.ReplaceAll(route, baseName, "")

	var completed string
	if len(route) > 0 {
		completed = basePath + "/" + route
	} else {
		completed = basePath
	}

	return completed
}

func chainAuthorizations(handler RequestHandler, authorizers AuthorizationMap) RequestHandler {
	for key, function := range authorizers {
		handler = function(handler, key)
	}
	return handler
}
