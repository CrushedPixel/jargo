package jargo

type Controller struct {
	Resource   *Resource
	Actions    *Actions
	Middleware []HandlerFunc
}

func NewController(model interface{}) (*Controller, error) {
	r, err := NewResource(model)
	if err != nil {
		return nil, err
	}

	a := make(Actions)
	actions := &a

	controller := &Controller{
		Resource: r,
		Actions:  actions,
	}

	return controller, nil
}

func NewCRUDController(model interface{}) (*Controller, error) {
	controller, err := NewController(model)
	if err != nil {
		return nil, err
	}

	controller.Actions.SetShowAction(NewAction(DefaultShowResourceHandler))
	controller.Actions.SetIndexAction(NewAction(DefaultIndexResourceHandler))
	controller.Actions.SetCreateAction(NewAction(DefaultCreateResourceHandler))
	controller.Actions.SetUpdateAction(NewAction(DefaultUpdateResourceHandler))
	controller.Actions.SetDeleteAction(NewAction(DefaultDeleteResourceHandler))

	return controller, nil
}

func (c *Controller) Use(middleware HandlerFunc) {
	c.Middleware = append(c.Middleware, middleware)
}

func (c *Controller) initialize(app *Application) {
	c.Resource.CreateTable(app.DB)

	// register actions
	for k, v := range *c.Actions {
		app.Register(v.toEndpoint(c, k))
	}
}
