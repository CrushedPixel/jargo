package jargo

// A Controller is responsible for all
// Actions related to a specific Resource.
type Controller struct {
	resource   *Resource
	middleware []MiddlewareFunc

	indexAction  indexHandlerChain
	showAction   showHandlerChain
	createAction createHandlerChain
	updateAction updateHandlerChain
	deleteAction deleteHandlerChain
}

// NewCRUDController returns a new Controller for a Resource
// with the default JSON API-compliant
// Index, Show, Create, Update and Delete Actions.
func NewCRUDController(resource *Resource) *Controller {
	c := NewController(resource)
	c.SetIndexAction(DefaultIndexResourceHandler)
	c.SetShowAction(DefaultShowResourceHandler)
	c.SetCreateAction(DefaultCreateResourceHandler)
	c.SetUpdateAction(DefaultUpdateResourceHandler)
	c.SetDeleteAction(DefaultDeleteResourceHandler)
	return c
}

// NewController returns a new Controller for a Resource.
func NewController(resource *Resource) *Controller {
	return &Controller{
		resource: resource,
	}
}

// Use adds handler functions to be run
// before the Controller's action handlers.
func (c *Controller) Use(middleware ...MiddlewareFunc) {
	c.middleware = append(c.middleware, middleware...)
}

// SetIndexAction sets the Controller's Index Action.
func (c *Controller) SetIndexAction(handlers ...IndexHandlerFunc) {
	c.indexAction = handlers
}

// SetShowAction sets the Controller's Show Action.
func (c *Controller) SetShowAction(handlers ...ShowHandlerFunc) {
	c.showAction = handlers
}

// SetCreateAction sets the Controller's Create Action.
func (c *Controller) SetCreateAction(handlers ...CreateHandlerFunc) {
	c.createAction = handlers
}

// SetUpdateAction sets the Controller's Update Action.
func (c *Controller) SetUpdateAction(handlers ...UpdateHandlerFunc) {
	c.updateAction = handlers
}

// SetDeleteAction sets the Controller's Delete Action.
func (c *Controller) SetDeleteAction(handlers ...DeleteHandlerFunc) {
	c.deleteAction = handlers
}
