package jargo

import "errors"

var errNoResponse = errors.New("the last HandlerFunc returned a nil Response")

// A HandlerFunc is a function handling a request.
type HandlerFunc func(context *Context, request Request) Response

// A Controller is responsible for all
// Actions related to a specific Resource.
type Controller struct {
	resource   *Resource
	middleware []HandlerFunc
	actions    map[ActionType][]HandlerFunc
}

// NewCRUDController returns a new Controller for a Resource
// with the default JSON API-compliant
// Index, Show, Create, Update and Delete Actions.
func NewCRUDController(resource *Resource) *Controller {
	c := NewController(resource)
	c.SetShowAction(DefaultShowResourceHandler)
	c.SetIndexAction(DefaultIndexResourceHandler)
	c.SetCreateAction(DefaultCreateResourceHandler)
	c.SetUpdateAction(DefaultUpdateResourceHandler)
	c.SetDeleteAction(DefaultDeleteResourceHandler)
	return c
}

// NewController returns a new Controller for a Resource.
func NewController(resource *Resource) *Controller {
	return &Controller{
		resource: resource,
		actions:  make(map[ActionType][]HandlerFunc),
	}
}

func (c *Controller) Handle(context *Context, request Request) Response {
	handlers, ok := c.actions[request.ActionType()]
	if !ok {
		return ErrNotFound // TODO
	}

	allHandlers := append(c.middleware, handlers...)

	var response Response
	for _, handler := range allHandlers {
		response = handler(context, request)
		if response != nil {
			return response
		}
	}

	return ErrorResponse(errNoResponse)
}

// Use adds handler functions to be run
// before the Controller's action handlers.
func (c *Controller) Use(middleware ...HandlerFunc) {
	c.middleware = append(c.middleware, middleware...)
}

// SetAction sets the Controller's action handlers
// for a given action type.
// If no handlers are provided, the action handlers
// for the action type are cleared.
func (c *Controller) SetAction(actionType ActionType, handlers ...HandlerFunc) {
	if len(handlers) > 0 {
		c.actions[actionType] = handlers
	} else {
		delete(c.actions, actionType)
	}
}

// SetIndexAction sets the Controller's Index Action.
// Shortcut for SetAction(jargo.ActionTypeIndex, handlers...)
func (c *Controller) SetIndexAction(handlers ...HandlerFunc) {
	c.SetAction(ActionTypeIndex, handlers...)
}

// SetShowAction sets the Controller's Show Action.
// Shortcut for SetAction(jargo.ActionTypeShow, handlers...)
func (c *Controller) SetShowAction(handlers ...HandlerFunc) {
	c.SetAction(ActionTypeShow, handlers...)
}

// SetCreateAction sets the Controller's Create Action.
// Shortcut for SetAction(jargo.ActionTypeCreate, handlers...)
func (c *Controller) SetCreateAction(handlers ...HandlerFunc) {
	c.SetAction(ActionTypeCreate, handlers...)
}

// SetUpdateAction sets the Controller's Update Action.
// Shortcut for SetAction(jargo.ActionTypeUpdate, handlers...)
func (c *Controller) SetUpdateAction(handlers ...HandlerFunc) {
	c.SetAction(ActionTypeUpdate, handlers...)
}

// SetDeleteAction sets the Controller's Delete Action.
// Shortcut for SetAction(jargo.ActionTypeDelete, handlers...)
func (c *Controller) SetDeleteAction(handlers ...HandlerFunc) {
	c.SetAction(ActionTypeDelete, handlers...)
}
