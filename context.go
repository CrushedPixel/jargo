package jargo

import (
	"github.com/crushedpixel/jargo/api"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
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

// Application returns the Application the Context belongs to.
func (c *Context) Application() *Application {
	a, _ := c.Get(keyApplication)
	return a.(*Application)
}

func (c *Context) setApplication(app *Application) {
	c.Set(keyApplication, app)
}

// Controller returns the Controller the Context belongs to.
func (c *Context) Controller() *Controller {
	b, _ := c.Get(keyController)
	return b.(*Controller)
}

func (c *Context) setController(cont *Controller) {
	c.Set(keyController, cont)
}

// Resource returns the Resource of the Controller the Context belongs to.
// Shortcut for Controller().Resource().
func (c *Context) Resource() api.Resource {
	return c.Controller().Resource()
}

// DB returns the DB handle of the Application the Context belongs to.
// Shortcut for Application().DB().
func (c *Context) DB() *pg.DB {
	return c.Application().DB()
}

// ResourceId returns the id path parameter as used
// by show, update and delete routes.
// Returns ErrInvalidId if the id is malformed.
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

// CreateModel parses the create request payload
// according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
// It returns a resource model instance.
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

// CreateModel parses the update request payload
// according to the JSON API spec.
// http://jsonapi.org/format/#crud-update
// It returns a resource model instance
// with the values updated as proposed in the request.
func (c *Context) UpdateModel() (interface{}, error) {
	m, ok := c.Get(keyUpdateModel)
	if !ok {
		var err error
		m, err = parseUpdateRequest(c)
		if err != nil {
			if _, ok := err.(*api.Error); ok {
				return nil, err
			}
			return nil, api.ErrInvalidPayload(err.Error())
		}
		c.Set(keyUpdateModel, m)
	}
	return m, nil
}

// Filters parses the request filter parameters
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching-filtering
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

// Filters parses the request field parameters
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching-sparse-fieldsets
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

// SortFields parses the request sort parameters
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching-sorting
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

// SortFields parses the request page parameters
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching-pagination
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

func parseCreateRequest(c *Context) (interface{}, error) {
	return c.Resource().ParseJsonapiPayload(c.Request.Body, c.Application().Validate())
}

func parseUpdateRequest(c *Context) (interface{}, error) {
	id, err := c.ResourceId()
	if err != nil {
		return nil, err
	}

	instance, err := c.Resource().SelectById(c.DB(), id).Result()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, api.ErrNotFound
		}
		return nil, err
	}

	return c.Resource().ParseJsonapiUpdatePayload(c.Request.Body, instance, c.Application().Validate())
}
