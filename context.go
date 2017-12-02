package jargo

import (
	"crushedpixel.net/margo"
)

const (
	application = "__jargoApplication"
	controller  = "__jargoController"

	fetchParams = "__fetchParams"
)

type Context struct {
	*margo.Context
}

// JSON API query parameters for fetching data
// (see http://jsonapi.org/format/#fetching)
// Custom filter mechanism: filter[name:like]=*name*
type FetchParams struct {
	// Include params.Include
	// Fields  params.Fields
	Filter Filters
	Sort   Sorting
	Page   Pagination
}

// set by injectApplicationMiddleware
func (c *Context) GetApplication() *Application {
	a, _ := c.Get(application)
	return a.(*Application)
}

// set in Action#toMargoHandler
func (c *Context) GetController() *Controller {
	b, _ := c.Get(controller)
	return b.(*Controller)
}

// set by fetchParamsMiddleware
func (c *Context) GetFetchParams() *FetchParams {
	p, _ := c.Get(fetchParams)
	return p.(*FetchParams)
}
