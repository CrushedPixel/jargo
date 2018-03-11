package jargo

// A Controller is responsible for all
// Actions related to a specific Resource.
type Controller struct {
	resource   *Resource
	middleware handlerChain

	indexHandlers  indexHandlerChain
	showHandlers   showHandlerChain
	createHandlers createHandlerChain
	updateHandlers updateHandlerChain
	deleteHandlers deleteHandlerChain

	customHandlers map[route]handlerChain
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
	c.SetCreateHandler(NewCreateAction())
	c.SetUpdateHandler(NewUpdateAction())
	c.SetDeleteHandler(NewDeleteAction())
	return c
}

// NewController creates a new controller for a given resource.
// If the Application already has a controller for this resource,
// it is replaced with the newly created controller.
func (app *Application) NewController(resource *Resource) *Controller {
	c := &Controller{
		resource:       resource,
		customHandlers: make(map[route]handlerChain),
	}

	app.controllers[c.resource] = c
	return c
}

// Use adds handlers to be executed
// before the Controller's action handlers.
func (c *Controller) Use(middleware ...Handler) {
	c.middleware = append(c.middleware, middleware...)
}

// UseFunc is a convenience method for Use,
// allowing the use of function literals without
// casting them to HandlerFunc.
func (c *Controller) UseFunc(middleware ...HandlerFunc) {
	for _, m := range middleware {
		c.Use(m)
	}
}

// SetIndexHandler sets the Controller's index request handlers.
func (c *Controller) SetIndexHandler(handlers ...IndexHandler) {
	c.indexHandlers = handlers
}

// SetIndexHandlerFunc is a convenience method for SetIndexHandler,
// allowing the use of function literals without
// casting them to IndexHandlerFunc.
func (c *Controller) SetIndexHandlerFunc(handlers ...IndexHandlerFunc) {
	c.indexHandlers = nil
	for _, h := range handlers {
		c.indexHandlers = append(c.indexHandlers, h)
	}
}

// SetShowHandler sets the Controller's show request handler.
func (c *Controller) SetShowHandler(handlers ...ShowHandler) {
	c.showHandlers = handlers
}

// SetShowHandlerFunc is a convenience method for SetShowHandler,
// allowing the use of function literals without
// casting them to ShowHandlerFunc.
func (c *Controller) SetShowHandlerFunc(handlers ...ShowHandlerFunc) {
	c.showHandlers = nil
	for _, h := range handlers {
		c.showHandlers = append(c.showHandlers, h)
	}
}

// SetCreateHandler sets the Controller's create request handler.
func (c *Controller) SetCreateHandler(handlers ...CreateHandler) {
	c.createHandlers = handlers
}

// SetCreateHandlerFunc is a convenience method for SetCreateHandler,
// allowing the use of function literals without
// casting them to CreateHandlerFunc.
func (c *Controller) SetCreateHandlerFunc(handlers ...CreateHandlerFunc) {
	c.createHandlers = nil
	for _, h := range handlers {
		c.createHandlers = append(c.createHandlers, h)
	}
}

// SetUpdateHandler sets the Controller's update request handler.
func (c *Controller) SetUpdateHandler(handlers ...UpdateHandler) {
	c.updateHandlers = handlers
}

// SetUpdateHandlerFunc is a convenience method for SetUpdateHandler,
// allowing the use of function literals without
// casting them to UpdateHandlerFunc.
func (c *Controller) SetUpdateHandlerFunc(handlers ...UpdateHandlerFunc) {
	c.updateHandlers = nil
	for _, h := range handlers {
		c.updateHandlers = append(c.updateHandlers, h)
	}
}

// SetDeleteHandler sets the Controller's delete request handler.
func (c *Controller) SetDeleteHandler(handlers ...DeleteHandler) {
	c.deleteHandlers = handlers
}

// SetDeleteHandlerFunc is a convenience method for SetDeleteHandler,
// allowing the use of function literals without
// casting them to DeleteHandlerFunc.
func (c *Controller) SetDeleteHandlerFunc(handlers ...DeleteHandlerFunc) {
	c.deleteHandlers = nil
	for _, h := range handlers {
		c.deleteHandlers = append(c.deleteHandlers, h)
	}
}

// SetHandler sets the Controller's handler for a given method and path.
func (c *Controller) SetHandler(method string, path string, handlers ...Handler) {
	// ensure leading slash in path unless path is empty
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	c.customHandlers[route{method: method, path: path}] = handlers
}

// SetHandlerFunc is a convenience method for SetHandler,
// allowing the use of function literals without
// casting them to HandlerFunc.
func (c *Controller) SetHandlerFunc(method string, path string, handlers ...HandlerFunc) {
	var h []Handler
	for _, handler := range handlers {
		h = append(h, handler)
	}
	c.SetHandler(method, path, h...)
}
