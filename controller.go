package jargo

import "crushedpixel.net/jargo/api"

type Controller struct {
	Namespace  string
	Resource   api.Resource
	Actions    Actions
	Middleware []HandlerFunc
}

func NewController(namespace string, resource api.Resource) *Controller {
	actions := make(Actions)

	controller := &Controller{
		Namespace: namespace,
		Resource:  resource,
		Actions:   actions,
	}

	return controller
}

func NewRootController(resource api.Resource) *Controller {
	return NewController("", resource)
}

func NewCRUDController(resource api.Resource) *Controller {
	controller := NewRootController(resource)

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
	// register actions
	for k, v := range c.Actions {
		app.Register(v.toEndpoint(c, k))
	}
}
