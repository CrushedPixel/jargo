package jargo

import (
	"crushedpixel.net/margo"
	"crushedpixel.net/jargo/api"
)

var DefaultIndexResourceHandler HandlerFunc = func(c *Context) margo.Response {
	filters, err := c.Filters()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	fields, err := c.FieldSet()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	sort, err := c.SortFields()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	pagination, err := c.Pagination()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	return c.Resource().Select(c.Application().DB).
		Filters(filters).
		Fields(fields).
		Sort(sort).
		Pagination(pagination)
}

var DefaultShowResourceHandler HandlerFunc = func(c *Context) margo.Response {
	fields, err := c.FieldSet()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	id, err := c.ResourceId()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	return c.Resource().SelectById(c.Application().DB, id).
		Fields(fields)
}

var DefaultCreateResourceHandler HandlerFunc = func(c *Context) margo.Response {
	fields, err := c.FieldSet()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	m, err := c.CreateModel()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	return c.Resource().InsertOne(c.Application().DB, m).
		Fields(fields)
}

var DefaultUpdateResourceHandler HandlerFunc = func(c *Context) margo.Response {
	fields, err := c.FieldSet()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	m, err := c.UpdateModel()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	return c.Resource().UpdateOne(c.Application().DB, m).
		Fields(fields)
}

var DefaultDeleteResourceHandler HandlerFunc = func(c *Context) margo.Response {
	id, err := c.ResourceId()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	return c.Resource().DeleteById(c.Application().DB, id)
}
