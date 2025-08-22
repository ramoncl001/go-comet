package main

import (
	"fmt"

	"github.com/ramoncl001/go-comet/comet"
)

type TaskController struct {
	comet.ControllerBase
}

func (TaskController) Get(r *comet.Request) comet.Response {
	return comet.Ok("Hello task")
}

func (TaskController) Post(r *comet.Request) comet.Response {
	return comet.Created("Tremendo post")
}

func (TaskController) Route() string {
	return ""
}

func (TaskController) Policies() comet.PoliciesConfig {
	return comet.PoliciesConfig{
		"Get": []comet.Policy{
			{
				Validation: func(next comet.RequestHandler, val interface{}) comet.RequestHandler {
					return func(r *comet.Request) comet.Response {
						fmt.Println("Validacion perro")
						return next(r)
					}
				},
				Value: "Hello",
			},
		},
	}
}

func main() {
	router := comet.NewDefaultRouter()

	router.Use(func(next comet.RequestHandler) comet.RequestHandler {
		return func(r *comet.Request) comet.Response {
			fmt.Println("Middleware global")
			return next(r)
		}
	})

	group := comet.Group("/person")

	group.MapGet("", func(r *comet.Request) comet.Response {
		return comet.Ok("Hello world")
	})

	group.MapGet("/:id", func(r *comet.Request) comet.Response {
		id := r.PathParams["id"]
		return comet.Ok("This is the id " + id)
	}, func(next comet.RequestHandler) comet.RequestHandler {
		return func(r *comet.Request) comet.Response {
			fmt.Println("Middleware personal")
			return next(r)
		}
	})

	router.MapController(TaskController{})

	router.MapGroup(group)

	err := router.Run()
	if err != nil {
		fmt.Print("ERROR")
	}
}
