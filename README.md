# Comet Framework

#### A modular based and eficient framework for web apps building

## Table of content
* [Requirements](#requirements)
* [Quickstart](#quickstart)
* [Installation](#installation)
* [Router](#router)
    - [Definition](#router-definition)
    - [Usage](#router-usage)
* [Groups](#groups)
    - [Definition](#groups-definition)
    - [Usage](#groups-usage)
* [Controllers](#controllers)
    - [Defining routes](#defining-routes)
    - [Defining path parameters](#defining-path-parameters)
    - [Mapping](#mapping)
* [Middlewares](#middlewares)
    - [Basic Examples](#basic-examples)
* [Dependency injection](#dependency-injection)

## Requirements

`go: v1.18+`

## Quickstart

### Comet-CLI

In order to create a new `comet` project we first need to install `comet-cli` with the following command:

```bash
go install github.com/ramoncl001/comet-cli@latest
```

Once `comet-cli` is installed we can create a new project with the following command:

```bash
comet-cli new <project-name> <module-name>
```

This execution will create a folder name after your project name with the following structure:

```text
your-project
├── go.mod
├── go.sum
├── infrastructure
├── main.go
├── middlewares
│   └── your_middleware.go
└── modules
    └── foo
        ├── controllers
        │   └── foo_controller.go
        ├── domain
        └── services
            └── foo_service.go
```

After the project creation type:

```bash
cd your-project
go run .
```

You should get a result like this:

```bash
Running server in :8080...
Routes:
```

## Manual Installation

To install comet in your project you can run:
```bash
go get -u github.com/ramoncl001/go-comet/comet@latest
```

## Router
Comet router is the core entity wich will manage all application endpoints and middlewares

### Router Definition
In order to create a new Comet Router we can use two ways. The default, wich works in `localhost:5051` and a custom version:

```go
// Default router
router := comet.NewDefaultRouter()

// Custom one
router := &comet.Router{
    Address: ":8080"
}
```

### Router Usage
The router is used to map application handlers and start the api server, let's see a basic example:

```go
router := comet.NewDefaultRouter()

router.MapGet("", func(r *comet.Request) comet.Response {
    return comet.Ok("Hello world")
})

if err := router.Run(); err != nil {
    panic(err)
}
```

In this basic example we have created a simple Hello World web application, wich handles a GET request at `http://localhost:5051/` and it returns `"Hello world"`

## Groups
In comet we can create groups to store a bunch of handler under a single base route

### Groups Definition
```go
group := comet.Group("/person")
```

### Groups usage
Groups can map many handlers and then they can be mapped into the router
```go
router := comet.NewDefaultRouter()

personGroup := comet.Group("/person")

personGroup.MapGet("", func(r *comet.Request) comet.Response {
    return comet.Ok("Here's the person")
})

router.MapGroup(personGroup)

_ = router.Run()
```

## Controllers
Comet can be used as a Controller-Based framework

To create a controller you can do it manually or using the cli typing the following command:

```bash
comet-cli add controller <name> <location>
```

After execution you should get a new file called after the name of your controller with this name structure `<name>_controller.go` and should look like this:

```bash
comet-cli add controller Person controllers
```

```go
package controllers

type PersonControler struct {
    comet.ControllerBase
}

func NewPersonController() PersonController {
    return PersonController{}
}

func (PersonController) Route() string {
    return ""
}

func (PersonController) Policies() comet.PoliciesConfig {
    return comet.PoliciesConfig{}
}
```

`Route()`: This method returns the controller route (eg. `/person`). If its empty it will return the controller name

`Policies()`: This method returns controller policies configuration for roles and permissions for every endpoint. You can see more in the [Policies](#policies) section

## Defining routes

Comet defines every endpoint route based in the controller name and the controller endpoint method names, let's see an example:

```go
func (PersonController) Get(r *comet.Request) comet.Response {
    return comet.Ok("Hello world")
} 
```

In this example we are defining a simple GET method in the Person controller, this will add a path: `[GET] - /person` in the router. The request method is defined by the function name prefix `List`, `Get`, `Post`, `Put`, `Patch`, `Delete`

## Defining path parameters

In comet the path parameters are identified with a group of keywords `By`, `With`, `For`, `Of`

### Examples:

```go
// [GET] - person/:id
func (PersonController) GetByID(r *comet.Request) comet.Response {
    id := r.PathParams["id"]
    ...
}

// GET - person/:name
func (PersonController) ListForCountry(r *comet.Request) comet.Response {
    name := r.PathParams["name"]
    ...
}

// POST - person/create
func (PersonController) PostCreate(r *comet.Request) comet.Response {
    ...
}

// PUT - person/update/:id
func (PersonController) PutUpdateByID(r *comet.Request) comet.Response {
    ...
}

// DELETE - person/:id
func (PersonController) DeleteByID(r *comet.Request) comet.Response (
    ...
)
```

### Mapping

Once the controller is created it can be mapped in the app just like this:

```go
router := comet.NewDefaulRouter()

controller := NewPersonController()
router.MapController(controller)

_ := router.Run()
```

## Middlewares
Comet have a custom definition for middlewares
```go
type Middleware = func(next comet.RequestHandler) comet.RequestHandler
```

This middlewares are executed before or after the execution of one or more handlers in the application

### Basic Examples

In order to build a simple middleware to print a request's content we can do it like this:
```go
var logMiddleware = func(next comet.RequestHandler) comet.RequestHandler {
    return func(r *comet.Request) comet.Response {
        fmt.Printf("Request: %v", r)
        return next(r)
    }
}

router := comet.NewDefaultRouter()

router.Use(logMiddleware)

router.MapGet("", func(r *comet.Request) comet.Response {
    return comet.NoContent()
})

_ = router.Run()
```

You can also add a middleware to a single group of handlers

```go
var logMiddleware = func(next comet.RequestHandler) comet.RequestHandler {
    return func(r *comet.Request) comet.Response {
        fmt.Printf("Request: %v", r)
        return next(r)
    }
}

group := comet.Group("v1/foo")
group.Use(logMiddleware)
```

Or in a simple way you can just add it to a single handler

```go

var logMiddleware = func(next comet.RequestHandler) comet.RequestHandler {
    return func(r *comet.Request) comet.Response {
        fmt.Printf("Request: %v", r)
        return next(r)
    }
}

var requestHandler = func(r *comet.Request) comet.Response {
    return comet.NoContent()
}

router.MapGet("", requestHandler, logMiddleware)
```

## Dependency injection
Like other frameworks and libraries like ASP.NET or Spring, Comet also have dependency injection support. In order to register some service or dependency we will have two choices, we can use regular registration or keyed registration, wich give us the posibility of register many instances of a service under a single interface without overwrite the already registered service. Here are some examples of Dependency Injection in Comet:

### Installing
Dependency injection is not built-in comet basic functionalities, you have to add it manually by installing it:

```bash
go get -u github.com/ramoncl001/go-comet/ioc@latest
```

### Scoped

Scoped dependencies life cycle starts when they are first resolved, and ends when the given context is closed.

```go
// Definition
func RegisterScoped[T any](provider interface{})
```

`T`: Type of the dependency interface

`provider`: Constructor function for the dependency instance


```go
type ServiceA interface {}

type serviceA struct {
    ServiceA
}

func NewServiceA() ServiceA {
    return &serviceA{}
}

...

// Registering the dependency
comet.RegisterScoped[ServiceA](NewServiceA)
```

### Keyed Scoped

Scoped dependencies life cycle starts when they are first resolved, and ends when the given context is closed.

```go
// Definition
func RegisterKeyedScoped[T any](provider interface{}, key interface{})
```

`T`: Type of the dependency interface

`provider`: Constructor function for the dependency instance

`key`: Key for dependency mapping

```go
type ServiceA interface {}

type serviceA struct {
    ServiceA
}

func NewServiceA() ServiceA {
    return &serviceA{}
}

type secondServiceA struct {
    ServiceA
}

func NewSecondServiceA ServiceA {
    return &secondServiceA{}
}

...

// Registering the dependencies
comet.RegisterKeyedScoped[ServiceA](NewServiceA, 1)

comet.RegisterKeyedScoped[ServiceA](NewSecondServiceA, 2)
```

### Transient

Transient dependencies life cycle starts once they are resolved and ends when the code block they where resolved to ends.

```go
// Definition
func RegisterTransient[T any](provider interface{})
```

`T`: Type of the dependency interface

`provider`: Constructor function for the dependency instance

```go
type ServiceA interface {}

type serviceA struct {
    ServiceA
}

func NewServiceA() ServiceA {
    return &serviceA{}
}

...

// Registering the dependency
comet.RegisterTransient[ServiceA](NewServiceA)
```


### Keyed Transient

Transient dependencies life cycle starts once they are resolved and ends when the code block they where resolved to ends.

```go
// Definition
func RegisterKeyedTransient[T any](provider interface{}, key interface{})
```

`T`: Type of the dependency interface

`provider`: Constructor function for the dependency instance

`key`: Key for dependency mapping

```go
type ServiceA interface {}

type serviceA struct {
    ServiceA
}

func NewServiceA() ServiceA {
    return &serviceA{}
}

type secondServiceA struct {
    ServiceA
}

func NewSecondServiceA ServiceA {
    return &secondServiceA{}
}

...

// Registering the dependencies
comet.RegisterKeyedTransient[ServiceA](NewServiceA, 1)

comet.RegisterKeyedTransient[ServiceA](NewSecondServiceA, 2)
```

### Singleton

Singleton dependencies lifecylce starts once they are registered and ends once the application is stopped.

```go
func RegisterSingleton[T any](instance T)
```

`T`: The literal type of the dependency

`instance`: An instance of the dependency (It will live until the application is closed)

```go
type ServiceA struct {
    Count int
}

func (s *ServiceA) RaiseCount() int {
    return s.Count++
}
...

// Register dependency
comet.RegisterSingleton(&ServiceA{Count: 0})
```

### Keyed Singleton

Singleton dependencies lifecylce starts once they are registered and ends once the application is stopped.

```go
func RegisterKeyedSingleton[T any](instance T, key interface{})
```

`T`: The literal type of the dependency

`instance`: An instance of the dependency (It will live until the application is closed)

`key`: Key for dependency mapping

```go

type CountService interface {
    RaiseCount() int
}

type ServiceA struct {
    CountService
    Count int
}

func (s *ServiceA) RaiseCount() int {
    return s.Count++
}

type ServiceB struct {
    CountService
    Count int
}

func (s *ServiceB) RaiseCount() int {
    return s.Count++
}

...

// Register dependency
comet.RegisterKeyedSingleton[CountService](&ServiceA{Count: 0}, 'A')
comet.RegisterKeyedSingleton[CountService](&ServiceB{Count: 0}, 'B')
```

### Resolve
After registering a service in the IoC container you can resolve it using the current context

```go
serviceA, err := ioc.Resolve[ServiceA](context.Background())
```

Even if you can use the background context it is recommended to use the request context, for example

```go
var handler = func(r *comet.Request) comet.Request {
    serviceA, err := ioc.Resolve[ServiceA](r.Context())
    result := serviceA.DoStuff()
    return comet.Ok(result)
}
```

### Resolve Keyed
Just like the regular `Resolve`, `KeyedResolve` returns an instance of a registered service, but this method allows to search this dependency by using a `key` previously defined in te registration.

```go
serviceA, err := ioc.ResolveKeyed[ServiceBase](context.Background(), 'A')

serviceB, err := ioc.ResolveKeyed[ServiceBase](context.Background(), 'B')
```
