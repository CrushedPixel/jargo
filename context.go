package jargo

import (
	"crushedpixel.net/margo"
	"github.com/gin-gonic/gin"
)

const (
	keyApplication  = "__jargoApplication"
	keyController   = "__jargoController"
	keyFetchParams  = "__fetchParams"
	keyCreatedModel = "__createdModel"
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

func (c *Context) GetApplication() *Application {
	a, _ := c.Get(keyApplication)
	return a.(*Application)
}

func setApplication(c *gin.Context, a *Application) {
	c.Set(keyApplication, a)
}

func (c *Context) GetController() *Controller {
	b, _ := c.Get(keyController)
	return b.(*Controller)
}

func (c *Context) setController(cont *Controller) {
	c.Set(keyController, cont)
}

func (c *Context) GetFetchParams() *FetchParams {
	p, _ := c.Get(keyFetchParams)
	return p.(*FetchParams)
}

func (c *Context) setFetchParams(p *FetchParams) {
	c.Set(keyFetchParams, p)
}

func (c *Context) GetCreatedModel() interface{} {
	m, _ := c.Get(keyCreatedModel)
	return m
}

func (c *Context) setCreatedModel(m interface{}) {
	c.Set(keyCreatedModel, m)
}
