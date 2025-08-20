package comet

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/ramoncl001/comet/ioc"
	"github.com/ramoncl001/comet/rest"
	"gorm.io/gorm"
)

// ApiServer represents the main API server instance.
// Handles HTTP requests, routing, middleware, and application lifecycle.
type ApiServer interface {
	MapController(controller interface{})
	UseDatabaseContext(dialector gorm.Dialector, args ...gorm.Option)
	UseMiddleware(m Middleware)
	AddJWTAuthentication(mg interface{}, provider JwtProvider, config JwtConfigurations, userConfig UserConfig)
	//UseAuthorization()
	//UseAuthentication()
	Run(addr string) error
}

// NewServer creates and returns a new instance of the API server.
// This is the entry point for initializing the Comet framework application.
func NewServer() ApiServer {
	return createServer()
}

type apiServer struct {
	ApiServer
	server      *http.ServeMux
	router      *router
	middlewares []Middleware
}

func createServer() ApiServer {
	return &apiServer{
		server: http.NewServeMux(),
		router: newRouter(),
	}
}

func (srv *apiServer) AddJWTAuthentication(mg interface{}, provider JwtProvider, config JwtConfigurations, userConfig UserConfig) {
	managerType := reflect.TypeOf(mg)
	if managerType.Kind() != reflect.Func {
		panic(managerType.Name() + "is not a SessionManager constructor function")
	}

	ioc.RegisterSingleton(&userConfig)
	ioc.RegisterTransient[UserManager](NewDefaultUserManager)
	ioc.RegisterSingleton(provider)
	ioc.RegisterSingleton(config)
	ioc.RegisterTransient[SessionManager](mg)
}

func (srv *apiServer) MapController(controller interface{}) {
	typ := reflect.TypeOf(controller).Out(0)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	name := typ.Name()
	ioc.RegisterKeyedScoped[ControllerBase](controller, name)
	ctrl, err := ioc.ResolveKeyedScoped[ControllerBase](context.Background(), typ.Name())
	if err != nil {
		panic(err)
	}

	srv.router.register(ctrl)
}

func (srv *apiServer) UseMiddleware(m Middleware) {
	srv.middlewares = append(srv.middlewares, m)
}

func (stc *apiServer) UseDatabaseContext(dialector gorm.Dialector, args ...gorm.Option) {
	ctx := NewDatabaseContext(dialector, args...)
	ioc.RegisterSingleton(ctx)
}

func (srv *apiServer) Run(addr string) error {
	fmt.Printf("Running server in %s...\n", addr)
	fmt.Println("Routes:")

	for _, controller := range srv.router.controllers {
		fmt.Printf("[%s]\n", controller.name)
		for _, route := range controller.dynamicRoutes {
			fmt.Printf("[%s]: %s\n", route.Method, route.PathPattern)
		}

		for key := range controller.staticRoutes {
			method := strings.Split(key, ":")[0]
			path := strings.Replace(key, method+":", "", 1)
			fmt.Printf("[%s]: %s\n", method, path+"/")
		}
	}

	middlewares := chain(srv.router.Handle, srv.middlewares...)

	srv.server.Handle("/", HTTPAdapter(middlewares))
	return http.ListenAndServe(addr, srv.server)
}

func getRouteName(baseName string) string {
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

func chain(handler RequestHandler, middlewares ...Middleware) RequestHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

type routeHandler struct {
	Function RequestHandler
	Name     string
}

type route struct {
	Method      string
	PathPattern string
	Handler     routeHandler
	PathParts   []string
	ParamNames  []string
}

type controller struct {
	name          string
	basePath      string
	staticRoutes  map[string]*routeHandler
	dynamicRoutes []route
}

type router struct {
	controllers map[string]controller
}

func (r *router) register(ctrl ControllerBase) {
	controllerType := reflect.TypeOf(ctrl)

	controller := controller{
		staticRoutes:  make(map[string]*routeHandler, 0),
		dynamicRoutes: make([]route, 0),
	}

	basePath := ""
	if ctrl.Route() == "" {
		basePath = getRouteName(controllerType.Name())
	} else {
		basePath = ctrl.Route()
	}

	controller.basePath = basePath
	name := strings.ReplaceAll(basePath, "/", "")

	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		if !isRequestMethod(method) {
			continue
		}

		invariantName := strings.ToUpper(method.Name)
		methodMap := []rest.RequestMethod{rest.GET, rest.POST, rest.DELETE, rest.PATCH, rest.POST, rest.PUT, rest.LIST}

		for _, prefix := range methodMap {
			if strings.HasPrefix(invariantName, prefix.String()) {
				path := getMethodPath(basePath, method.Name)
				httpMethod := prefix.Method()

				handler := func(req *Request) Response {
					n := controllerType.Name()
					ctrl, err := ioc.ResolveKeyedScoped[ControllerBase](req.Context(), n)
					if err != nil {
						return Error("error getting controller")
					}

					controller := reflect.ValueOf(ctrl)
					request := reflect.ValueOf(req)

					responses := method.Func.Call([]reflect.Value{controller, request})
					result := responses[0].Interface().(Response)
					return result
				}

				if !strings.Contains(path, ":") {
					key := fmt.Sprintf("%s:%s", httpMethod, path)
					controller.staticRoutes[key] = &routeHandler{
						Function: handler,
						Name:     method.Name,
					}
					break
				}

				parts := strings.Split(path, "/")
				params := make([]string, 0)

				for _, part := range parts {
					if strings.HasPrefix(part, ":") {
						params = append(params, part[1:])
					}
				}

				controller.dynamicRoutes = append(controller.dynamicRoutes, route{
					Method:      httpMethod,
					PathPattern: path,
					Handler: struct {
						Function RequestHandler
						Name     string
					}{
						Function: handler,
						Name:     method.Name,
					},
					PathParts:  parts,
					ParamNames: params,
				})

				break
			}
		}
	}

	controller.name = controllerType.Name()
	r.controllers[name] = controller
}

func (r *router) Handle(req *Request) Response {
	path := req.Url.Path

	name := strings.Split(path, "/")[1]
	controller := r.controllers[name]

	str := fmt.Sprintf("%s:%s", req.Method, path)

	var handler *routeHandler
	handler, ok := controller.staticRoutes[str]
	if !ok {
		for _, route := range controller.dynamicRoutes {
			if route.Method != req.Method {
				continue
			}

			if params, ok := r.matchPath(route, path); ok {
				req.PathParams = params
				handler = &route.Handler
			}
		}
	}

	if handler == nil {
		return NotFound()
	}

	ctrl, err := ioc.ResolveKeyedScoped[ControllerBase](req.Context(), controller.name)
	if err != nil {
		return NotFound()
	}

	authorizeMap := make(map[interface{}]AuthorizerFunction)
	config := make([]Policy, 0)

	globalPolicies := ctrl.Policies()["*"]
	if globalPolicies != nil {
		config = append(config, globalPolicies...)
	}

	handlerPolicies := ctrl.Policies()[handler.Name]
	if handlerPolicies != nil {
		config = append(config, handlerPolicies...)
	}

	for _, val := range config {
		authorizeMap[val.Value] = val.Validation
	}

	resultHandler := chainAuthorizations(handler.Function, authorizeMap)
	return resultHandler(req)
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

func newRouter() *router {
	return &router{
		controllers: make(map[string]controller),
	}
}

func isRequestMethod(method reflect.Method) bool {
	if method.Name[0] < 'A' || method.Name[0] > 'Z' {
		return false
	}

	validPrefixes := []rest.RequestMethod{rest.GET, rest.POST, rest.PUT, rest.PATCH, rest.DELETE, rest.LIST}

	methodName := strings.ToUpper(method.Name)
	matchPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(methodName, prefix.String()) {
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

	return reqType.Name() == "Request" && strings.Contains(reqType.PkgPath(), "github.com/ramoncl001/comet")
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
