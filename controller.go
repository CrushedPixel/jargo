package jargo

// A Controller is responsible for all
// Actions related to a specific Resource.
type Controller struct {
	resource   *Resource
	middleware handlerChain

	indexAction  indexHandlerChain
	showAction   showHandlerChain
	createAction createHandlerChain
	updateAction updateHandlerChain
	deleteAction deleteHandlerChain

	customActions map[route]handlerChain
}

type route struct {
	method string
	path   string
}

// NewCRUDController returns a new Controller for a Resource
// with the default JSON API-compliant
// Index, Show, Create, Update and Delete Actions.
// If the Application already has a controller for this resource,
// it is replaced with the newly created controller.
func (app *Application) NewCRUDController(resource *Resource) *Controller {
	c := app.NewController(resource)
	c.SetIndexAction(DefaultIndexResourceHandler)
	c.SetShowAction(DefaultShowResourceHandler)
	c.SetCreateAction(DefaultCreateResourceHandler)
	c.SetUpdateAction(DefaultUpdateResourceHandler)
	c.SetDeleteAction(DefaultDeleteResourceHandler)
	return c
}

// NewController creates a new controller for a given resource.
// If the Application already has a controller for this resource,
// it is replaced with the newly created controller.
func (app *Application) NewController(resource *Resource) *Controller {
	c := &Controller{
		resource:      resource,
		customActions: make(map[route]handlerChain),
	}

	app.controllers[c.resource] = c
	return c
}

// Use adds handler functions to be run
// before the Controller's action handlers.
func (c *Controller) Use(middleware ...Handler) {
	c.middleware = append(c.middleware, middleware...)
}

// SetIndexAction sets the Controller's Index Action.
func (c *Controller) SetIndexAction(handlers ...IndexHandler) {
	c.indexAction = handlers
}

// SetShowAction sets the Controller's Show Action.
func (c *Controller) SetShowAction(handlers ...ShowHandler) {
	c.showAction = handlers
}

// SetCreateAction sets the Controller's Create Action.
func (c *Controller) SetCreateAction(handlers ...CreateHandler) {
	c.createAction = handlers
}

// SetUpdateAction sets the Controller's Update Action.
func (c *Controller) SetUpdateAction(handlers ...UpdateHandler) {
	c.updateAction = handlers
}

// SetDeleteAction sets the Controller's Delete Action.
func (c *Controller) SetDeleteAction(handlers ...DeleteHandler) {
	c.deleteAction = handlers
}

func (c *Controller) SetAction(method string, path string, handlers ...Handler) {
	// ensure leading slash in path unless path is empty
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	c.customActions[route{method: method, path: path}] = handlers
}
