package jargo

import (
	"crushedpixel.net/margo"
	"crushedpixel.net/jargo/api"
)

func DefaultIndexResourceHandler(c *Context) margo.Response {
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

func DefaultShowResourceHandler(c *Context) margo.Response {
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

func DefaultCreateResourceHandler(c *Context) margo.Response {
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

func DefaultUpdateResourceHandler(c *Context) margo.Response {
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

func DefaultDeleteResourceHandler(c *Context) margo.Response {
	id, err := c.ResourceId()
	if err != nil {
		return api.NewErrorResponse(err)
	}

	return c.Resource().DeleteById(c.Application().DB, id)
}
