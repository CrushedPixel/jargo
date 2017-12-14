package jargo

import (
	"crushedpixel.net/margo"
	"github.com/gin-gonic/gin"
)

const (
	keyApplication      = "__jargoApplication"
	keyController       = "__jargoController"
	keyQueryParams      = "__queryParams"
	keyIndexQueryParams = "__indexQueryParams"
	keyCreateModel      = "__createdModel"
	keyUpdateModel      = "__updateModel"
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

func (c *Context) GetQueryParams() (*QueryParams, error) {
	p, ok := c.Get(keyQueryParams)
	if !ok {
		var err error
		p, err = parseQueryParams(c)
		if err != nil {
			return nil, invalidQueryParams(err)
		}
		c.Set(keyQueryParams, p)
	}
	return p.(*QueryParams), nil
}

func (c *Context) GetIndexQueryParams() (*IndexQueryParams, error) {
	p, ok := c.Get(keyIndexQueryParams)
	if !ok {
		var err error
		p, err = parseIndexQueryParams(c)
		if err != nil {
			return nil, invalidQueryParams(err)
		}

		c.Set(keyIndexQueryParams, p)
	}
	return p.(*IndexQueryParams), nil
}

func (c *Context) GetCreateModel() (interface{}, error) {
	m, ok := c.Get(keyCreateModel)
	if !ok {
		var err error
		m, err = parseCreateRequest(c)
		if err != nil {
			return nil, invalidPayload(err)
		}
		c.Set(keyCreateModel, m)
	}
	return m, nil
}

func (c *Context) GetUpdateModel() (interface{}, error) {
	m, ok := c.Get(keyUpdateModel)
	if !ok {
		var err error
		m, err = parseUpdateRequest(c)
		if err != nil {
			return nil, invalidPayload(err)
		}
		c.Set(keyUpdateModel, m)
	}
	return m, nil
}
