package jargo

import (
	"crushedpixel.net/margo"
	"github.com/gin-gonic/gin"
)

const (
	keyApplication = "__jargoApplication"
	keyController  = "__jargoController"
	keyFetchParams = "__fetchParams"
	keyCreateModel = "__createdModel"
)

type Context struct {
	*margo.Context
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
	p, ok := c.Get(keyFetchParams)
	if !ok {
		var err error
		p, err = parseFetchRequest(c)
		if err != nil {
			panic(invalidQueryParams(err))
		}
		c.Set(keyFetchParams, p)
	}
	return p.(*FetchParams)
}

func (c *Context) GetCreateModel() interface{} {
	m, ok := c.Get(keyCreateModel)
	if !ok {
		var err error
		m, err = parseCreateRequest(c)
		if err != nil {
			panic(invalidPayload(err))
		}
		c.Set(keyCreateModel, m)
	}
	return m
}
