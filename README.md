# Comet Framework

#### A modular based and eficient framework for web apps building

## Table of content
* [Requirements](#requirements)
* [Quickstart](#quickstart)
* [Installation](#installation)
* [Controllers](#controllers)
    - [Defining routes](#defining-routes)
    - [Defining path parameters](#defining-path-parameters)
    - [Mapping](#mapping)
* [Dependency injection](#dependency-injection)

## Requirements

`go: v1.18+`

## Quickstart

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
[FooController]
```

## Installation

To install comet in your project you can run:
```bash
go get -u github.com/ramoncl001/go-comet/comet@latest
```

## Controllers
Comet is Controller-Based framework

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

## Mapping

Once the controller is created it can be mapped in the app just like this:

```go
app := comet.NewServer()

app.MapController(controllers.NewPersonController)

app.Run(":8080")
```

## Dependency injection
Like other frameworks and libraries like ASP.NET or Spring, Comet also have dependency injection support. In order to register some service or dependency we will have two choices, we can use regular registration or keyed registration, wich give us the posibility of register many instances of a service under a single interface without overwrite the already registered service. Here are some examples of Dependency Injection in Comet:

### Scoped

Scoped dependencies life cycle starts when they are first resolved, and ends when the given context is closed.

```go
// Definition
func RegisterScoped[T any](provider interface{})
```

`T`: Type of the dependency interface

`provider`: Constructor function for the dependency instance

```go
func ResolveScoped[T any](ctx context.Context) (T, error)
```

`T`: Type of the dependency interface

`ctx`: Current context

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

// Resolving the dependency
ctx := context.Background()
service, err := comet.ResolveScoped[ServiceA](ctx)
if err != nil {
    panic("error resolving dependency")
}
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
func ResolveKeyedScoped[T any](ctx context.Context, key interface{}) (T, error)
```

`T`: Type of the dependency interface

`ctx`: Current context

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

// Resolving the dependencies
ctx := context.Background()

service, err := comet.ResolveKeyedScoped[ServiceA](ctx, 1)
if err != nil {
    panic("error resolving dependency")
}

secondService, err := comet.ResolveKeyedScoped[ServiceA](ctx, 2)
if err != nil {
    panic("error resolving dependency")
}
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
func ResolveTransient[T any](ctx context.Context) (T, error)
```

`T`: Type of the dependency interface

`ctx`: Current context

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

// Resolving the dependency
ctx := context.Background()
service, err := comet.ResolveTransient[ServiceA](ctx)
if err != nil {
    panic("error resolving dependency")
}
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
func ResolveKeyedTransient[T any](ctx context.Context, key interface{}) (T, error)
```

`T`: Type of the dependency interface

`ctx`: Current context

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

// Resolving the dependencies
ctx := context.Background()

service, err := comet.ResolveKeyedTransient[ServiceA](ctx, 1)
if err != nil {
    panic("error resolving dependency")
}

secondService, err := comet.ResolveKeyedTransient[ServiceA](ctx, 2)
if err != nil {
    panic("error resolving dependency")
}
```

### Singleton

Singleton dependencies lifecylce starts once they are registered and ends once the application is stopped.

```go
func RegisterSingleton[T any](instance T)
```

`T`: The literal type of the dependency

`instance`: An instance of the dependency (It will live until the application is closed)

```go
func ResolveSingleton[T any](ctx context.Context) (T, error)
```

`T`: Literal type of the dependency

`ctx`: Current context

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

// Retrieve dependency
ctx := context.Background()
service, err := comet.ResolveSingleton[*ServiceA](ctx)


// Count property will keep increasing until app shutdown
fmt.Print(service.RaiseCount())
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
func ResolveKeyedSingleton[T any](ctx context.Context, key interface{}) (T, error)
```

`T`: Literal type of the dependency

`ctx`: Current context

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

// Retrieve dependency
ctx := context.Background()
countA, err := comet.ResolveKeyedSingleton[CountService](ctx, 'A')

countB, err := comet.ResolveKeyedSingleton[CountService](ctx, 'B')

// Each count will increase by its own
fmt.Print(serviceA.RaiseCount())
fmt.Print(serviceB.RaiseCount())
```
