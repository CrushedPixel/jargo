package jargo

type Controller struct {
	Resource *Resource
	Actions  *Actions
}

func NewController(model interface{}, defaultActions bool) (*Controller, error) {
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

	if defaultActions {
		controller.Actions.SetShowAction(NewAction(DefaultShowResourceHandler))
		controller.Actions.SetIndexAction(NewAction(DefaultIndexResourceHandler))
		controller.Actions.SetCreateAction(NewAction(DefaultCreateResourceHandler))
	}
	return controller, nil
}

func (c *Controller) initialize(app *Application) {
	c.Resource.CreateTable(app.DB)

	// register actions
	for k, v := range *c.Actions {
		app.Register(v.toEndpoint(c, k))
	}
}
