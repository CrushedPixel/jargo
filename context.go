package jargo

import (
	"bytes"
	"github.com/go-pg/pg"
	"net/http"
	"strconv"
)

type Context struct {
	application *Application
	resource    *Resource

	resourceId int64

	fieldSet   *FieldSet
	filters    *Filters
	pagination *Pagination
	sort       *SortFields

	header  http.Header
	payload string

	data map[string]interface{}
}

func NewContext(app *Application, resource *Resource, req *Request) (*Context, error) {
	context := &Context{
		application: app,
		resource:    resource,
		data:        make(map[string]interface{}),
	}

	if req.ActionType == ActionTypeCreate ||
		req.ActionType == ActionTypeUpdate {
		// read payload into string
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(req.Payload)
		if err != nil {
			return nil, err
		}
		context.payload = buf.String()
	}

	// parse resource id
	if req.ActionType == ActionTypeShow ||
		req.ActionType == ActionTypeUpdate ||
		req.ActionType == ActionTypeDelete {
		id, err := strconv.ParseInt(req.ResourceId, 10, 0)
		if err != nil {
			return nil, ErrInvalidId
		}
		context.resourceId = id
	}

	// parse sparse fieldset
	if req.ActionType == ActionTypeIndex ||
		req.ActionType == ActionTypeShow ||
		req.ActionType == ActionTypeUpdate ||
		req.ActionType == ActionTypeDelete {
		fs, err := resource.ParseFieldSet(req.Fields)
		if err != nil {
			return nil, err
		}
		context.fieldSet = fs
	}

	if req.ActionType == ActionTypeIndex {
		// parse filters
		filters, err := resource.ParseFilters(req.Filters)
		if err != nil {
			return nil, err
		}
		context.filters = filters

		// parse pagination
		pagination, err := ParsePagination(req.Pagination, app.MaxPageSize())
		if err != nil {
			return nil, err
		}
		context.pagination = pagination

		// parse sort fields
		sort, err := resource.ParseSortFields(req.Sort)
		if err != nil {
			return nil, err
		}
		context.sort = sort
	}

	return context, nil
}

func (c *Context) Application() *Application {
	return c.application
}

func (c *Context) DB() *pg.DB {
	return c.application.db
}

func (c *Context) Resource() *Resource {
	return c.resource
}

func (c *Context) ResourceId() int64 {
	return c.resourceId
}

func (c *Context) FieldSet() *FieldSet {
	return c.fieldSet
}

func (c *Context) Filters() *Filters {
	return c.filters
}

func (c *Context) Pagination() *Pagination {
	return c.pagination
}

func (c *Context) SortFields() *SortFields {
	return c.sort
}

func (c *Context) Payload() string {
	return c.payload
}

func (c *Context) Header() http.Header {
	return c.header
}

func (c *Context) Get(key string) interface{} {
	return c.data[key]
}

func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}
