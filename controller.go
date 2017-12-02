package jargo

import (
	"crushedpixel.net/jargo/models"
)

type Controller struct {
	BasePath string
	Model    *models.Model
	Actions  *Actions
}

func NewController(path string, model interface{}, defaultActions bool) (*Controller, error) {
	m, err := models.New(model)
	if err != nil {
		return nil, err
	}

	a := make(Actions)
	actions := &a

	controller := &Controller{
		BasePath: path,
		Model:    m,
		Actions:  actions,
	}

	if defaultActions {
		controller.Actions.SetIndexAction(NewAction(func(c *Context) interface{} {
			return IndexResourceQuery(c)
		}))

		controller.Actions.SetCreateAction(NewAction(CreateResourceHandler))
	}

	return controller, nil
}

func (c *Controller) initialize(app *Application) {
	c.Model.CreateTable(app.DB)

	// register actions
	for k, v := range *c.Actions {
		app.Register(v.toEndpoint(c, k))
	}
}
