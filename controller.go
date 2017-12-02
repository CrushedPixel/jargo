package jargo

import (
	"crushedpixel.net/margo"
)

type Controller struct {
	BasePath string
	Model    *Model
	Actions  []*Action
}

func NewController(path string, model interface{}) (*Controller, error) {
	m, err := newModel(model)
	if err != nil {
		return nil, err
	}

	controller := &Controller{path, m, []*Action{}}
	//controller.AddAction(NewAction(http.MethodGet, "/", index))
	// TODO: builtin actions (index) and a way to override them

	return controller, nil
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
