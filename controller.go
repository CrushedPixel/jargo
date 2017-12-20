package jargo

import "crushedpixel.net/jargo/api"

type Controller struct {
	Resource   api.Resource
	Actions    *Actions
	Middleware []HandlerFunc
}

func NewController(resource api.Resource) *Controller {
	a := make(Actions)
	actions := &a

	controller := &Controller{
		Resource: resource,
		Actions:  actions,
	}

	return controller
}

func NewCRUDController(resource api.Resource) *Controller {
	controller := NewController(resource)

	controller.Actions.SetShowAction(NewAction(DefaultShowResourceHandler))
	controller.Actions.SetIndexAction(NewAction(DefaultIndexResourceHandler))
	controller.Actions.SetCreateAction(NewAction(DefaultCreateResourceHandler))
	controller.Actions.SetUpdateAction(NewAction(DefaultUpdateResourceHandler))
	controller.Actions.SetDeleteAction(NewAction(DefaultDeleteResourceHandler))

	return controller
}

func (c *Controller) Use(middleware HandlerFunc) {
	c.Middleware = append(c.Middleware, middleware)
}

func (c *Controller) initialize(app *Application) {
	c.Resource.CreateTable(app.DB)
	app.RegisterResource(c.Resource)

	// register actions
	for k, v := range *c.Actions {
		app.Register(v.toEndpoint(c, k))
	}
}
