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
	c.SetIndexHandler(NewIndexAction())
	c.SetShowHandler(NewShowAction())
	c.SetCreateHandler(DefaultCreateResourceHandler)
	c.SetUpdateHandler(DefaultUpdateResourceHandler)
	c.SetDeleteHandler(DefaultDeleteResourceHandler)
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

// SetIndexHandler sets the Controller's index request handler.
func (c *Controller) SetIndexHandler(handlers ...IndexHandler) {
	c.indexAction = handlers
}

// SetShowHandler sets the Controller's show request handler.
func (c *Controller) SetShowHandler(handlers ...ShowHandler) {
	c.showAction = handlers
}

// SetCreateHandler sets the Controller's create request handler.
func (c *Controller) SetCreateHandler(handlers ...CreateHandler) {
	c.createAction = handlers
}

// SetUpdateHandler sets the Controller's update request handler.
func (c *Controller) SetUpdateHandler(handlers ...UpdateHandler) {
	c.updateAction = handlers
}

// SetDeleteHandler sets the Controller's delete request handler.
func (c *Controller) SetDeleteHandler(handlers ...DeleteHandler) {
	c.deleteAction = handlers
}

// SetHandler sets the Controller's handler for a given method and path.
func (c *Controller) SetHandler(method string, path string, handlers ...Handler) {
	// ensure leading slash in path unless path is empty
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	c.customActions[route{method: method, path: path}] = handlers
}
