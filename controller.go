package jargo

import (
	"net/http"
)

// A Controller is responsible for all
// Actions related to a specific Resource.
type Controller struct {
	namespace  string
	resource   *Resource
	middleware []HandlerFunc
	actions    map[route]*Action
}

type route struct {
	method string
	path   string
}

var (
	showRoute   = route{http.MethodGet, "/:id"}
	indexRoute  = route{http.MethodGet, "/"}
	createRoute = route{http.MethodPost, "/"}
	updateRoute = route{http.MethodPatch, "/:id"}
	deleteRoute = route{http.MethodDelete, "/:id"}
)

// NewRootController returns a new Controller for a Resource at root level.
func NewRootController(resource *Resource) *Controller {
	return NewController("", resource)
}

// NewCRUDController returns a new Controller for a Resource
// and namespace with the default JSON API compliant
// index, show, create, update and delete Actions.
func NewCRUDController(namespace string, resource *Resource) *Controller {
	c := NewController(namespace, resource)
	c.SetShowAction(NewJSONAPIAction(DefaultShowResourceHandler))
	c.SetIndexAction(NewJSONAPIAction(DefaultIndexResourceHandler))
	c.SetCreateAction(NewJSONAPIAction(DefaultCreateResourceHandler))
	c.SetUpdateAction(NewJSONAPIAction(DefaultUpdateResourceHandler))
	c.SetDeleteAction(NewJSONAPIAction(DefaultDeleteResourceHandler))
	return c
}

// NewRootController returns a new Controller for a Resource and namespace.
// The namespace is prepended to the Route path of the Controller's Actions.
func NewController(namespace string, resource *Resource) *Controller {
	return &Controller{
		namespace: namespace,
		resource:  resource,
		actions:   make(map[route]*Action),
	}
}

// handlers returns a HandlerChain to be called
// when an Application handles a request for this Controller.
func (c *Controller) handlers(app *Application) HandlerChain {
	// prepend injectApplicationMiddleware
	// and injectControllerMiddleware
	// to Controller-level middleware
	handlers := append([]HandlerFunc{
		injectApplicationMiddleware(app),
		injectControllerMiddleware(c),
	}, c.middleware...)
	return HandlerChain(handlers)
}

// Namespace returns the Controller's namespace.
func (c *Controller) Namespace() string {
	return c.namespace
}

// Resource returns the Controller's Resource.
func (c *Controller) Resource() *Resource {
	return c.resource
}

// Use adds margo.HandlerFuncs to be run
// before any of the Controller's Actions.
func (c *Controller) Use(middleware ...HandlerFunc) {
	c.middleware = append(c.middleware, middleware...)
}

// Action returns the Controller's Action for a given route.
// May return nil.
func (c *Controller) Action(method string, path string) *Action {
	return c.actions[route{method, path}]
}

// SetAction sets the Controller's Action for a given route.
// Action can be nil.
func (c *Controller) SetAction(method string, path string, action *Action) {
	c.actions[route{method, path}] = action
}

// ShowAction returns the Controller's show Action.
// Shortcut for Action(http.MethodGet, "/:id")
func (c *Controller) ShowAction() *Action {
	return c.actions[showRoute]
}

// SetShowAction sets the Controller's show Action.
// Shortcut for SetAction(http.MethodGet, "/:id")
func (c *Controller) SetShowAction(action *Action) {
	c.actions[showRoute] = action
}

// IndexAction returns the Controller's index Action.
// Shortcut for Action(http.MethodGet, "/")
func (c *Controller) IndexAction() *Action {
	return c.actions[indexRoute]
}

// SetIndexAction sets the Controller's index Action.
// Shortcut for SetAction(http.MethodGet, "/")
func (c *Controller) SetIndexAction(action *Action) {
	c.actions[indexRoute] = action
}

// CreateAction returns the Controller's create Action.
// Shortcut for Action(http.MethodPost, "/")
func (c *Controller) CreateAction() *Action {
	return c.actions[createRoute]
}

// SetCreateAction sets the Controller's create Action.
// Shortcut for SetAction(http.MethodPost, "/")
func (c *Controller) SetCreateAction(action *Action) {
	c.actions[createRoute] = action
}

// UpdateAction returns the Controller's update Action.
// Shortcut for Action(http.MethodPatch, "/:id")
func (c *Controller) UpdateAction() *Action {
	return c.actions[updateRoute]
}

// SetUpdateAction sets the Controller's update Action.
// Shortcut for SetAction(http.MethodPatch, "/:id")
func (c *Controller) SetUpdateAction(action *Action) {
	c.actions[updateRoute] = action
}

// DeleteAction returns the Controller's delete Action.
// Shortcut for Action(http.MethodDelete, "/:id")
func (c *Controller) DeleteAction() *Action {
	return c.actions[deleteRoute]
}

// SetDeleteAction sets the Controller's delete Action.
// Shortcut for SetAction(http.MethodDelete, "/:id")
func (c *Controller) SetDeleteAction(action *Action) {
	c.actions[deleteRoute] = action
}
