package jargo

import (
	"crushedpixel.net/margo"
)

type Controller struct {
	BasePath string
	Model    interface{}
	Actions  []*Action
}

func NewController(path string, model interface{}) *Controller {
	controller := &Controller{path, model, []*Action{}}
	//controller.AddAction(NewAction(http.MethodGet, "/", index))
	// TODO: builtin actions (index) and a way to override them

	return controller
}

func (c *Controller) AddAction(action *Action) {
	c.Actions = append(c.Actions, action)
}

func (c *Controller) toEndpoints() []*margo.Endpoint {
	endpoints := make([]*margo.Endpoint, len(c.Actions))

	for i := range c.Actions {
		endpoints[i] = c.Actions[i].toEndpoint(c)
	}

	return endpoints
}

func index(c *Context) Response {
	app := c.GetApplication()
	cnt := c.GetController()

	app.DB.Model(cnt.Model).Select()

	// TODO: return database results, marshalled into jsonapi
	return NewDataResponse(200, nil)
}

// TODO utility functions for JSON API CRUDing
