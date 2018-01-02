package jargo

import (
	"github.com/gin-gonic/gin"
	"crushedpixel.net/jargo/api"
	"strconv"
)

const (
	keyApplication = "__jargoApplication"
	keyController  = "__jargoController"
	keyCreateModel = "__jargoCreatedModel"
	keyUpdateModel = "__jargoUpdateModel"

	keyFilters    = "__jargoFilters"
	keyFieldSet   = "__jargoFields"
	keySortFields = "__jargoSort"
	keyPagination = "__jargoPagination"

	keyResourceId = "__jargoResourceId"

	idParam = "id"
)

type Context struct {
	*gin.Context
}

func (c *Context) Application() *Application {
	a, _ := c.Get(keyApplication)
	return a.(*Application)
}

func setApplication(c *gin.Context, a *Application) {
	c.Set(keyApplication, a)
}

func (c *Context) Controller() *Controller {
	b, _ := c.Get(keyController)
	return b.(*Controller)
}

func (c *Context) setController(cont *Controller) {
	c.Set(keyController, cont)
}

func (c *Context) Resource() api.Resource {
	return c.Controller().Resource
}

func (c *Context) ResourceId() (int64, error) {
	id, ok := c.Get(keyResourceId)
	if !ok {
		var err error
		id, err = strconv.ParseInt(c.Param(idParam), 10, 0)
		if err != nil {
			return 0, api.ErrInvalidId
		}
		c.Set(keyResourceId, id)
	}
	return id.(int64), nil
}

func (c *Context) CreateModel() (interface{}, error) {
	m, ok := c.Get(keyCreateModel)
	if !ok {
		var err error
		m, err = parseCreateRequest(c)
		if err != nil {
			return nil, api.ErrInvalidPayload(err.Error())
		}
		c.Set(keyCreateModel, m)
	}
	return m, nil
}

func (c *Context) UpdateModel() (interface{}, error) {
	m, ok := c.Get(keyUpdateModel)
	if !ok {
		var err error
		m, err = parseUpdateRequest(c)
		if err != nil {
			if _, ok := err.(*api.ApiError); ok {
				return nil, err
			}
			return nil, api.ErrInvalidPayload(err.Error())
		}
		c.Set(keyUpdateModel, m)
	}
	return m, nil
}

func (c *Context) Filters() (api.Filters, error) {
	f, ok := c.Get(keyFilters)
	if !ok {
		var err error
		f, err = c.Resource().ParseFilters(c.Request.URL.Query())
		if err != nil {
			return nil, api.ErrInvalidQueryParams(err.Error())
		}
		c.Set(keyFilters, f)
	}
	return f.(api.Filters), nil
}

func (c *Context) FieldSet() (api.FieldSet, error) {
	f, ok := c.Get(keyFieldSet)
	if !ok {
		var err error
		f, err = c.Resource().ParseFieldSet(c.Request.URL.Query())
		if err != nil {
			return nil, api.ErrInvalidQueryParams(err.Error())
		}
		c.Set(keyFieldSet, f)
	}
	return f.(api.FieldSet), nil
}

func (c *Context) SortFields() (api.SortFields, error) {
	f, ok := c.Get(keySortFields)
	if !ok {
		var err error
		f, err = c.Resource().ParseSortFields(c.Request.URL.Query())
		if err != nil {
			return nil, api.ErrInvalidQueryParams(err.Error())
		}
		c.Set(keySortFields, f)
	}
	return f.(api.SortFields), nil
}

func (c *Context) Pagination() (api.Pagination, error) {
	f, ok := c.Get(keyPagination)
	if !ok {
		var err error
		f, err = c.Application().ParsePagination(c.Request.URL.Query())
		if err != nil {
			return nil, api.ErrInvalidQueryParams(err.Error())
		}
		c.Set(keyPagination, f)
	}
	return f.(api.Pagination), nil
}
