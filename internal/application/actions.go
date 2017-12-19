package application

import (
	"crushedpixel.net/margo"
)

var DefaultIndexResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	indexParams, err := c.GetIndexQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.GetController().Resource.Select(c.GetApplication().DB).
		ApplyQueryParams(params).
		ApplyIndexQueryParams(indexParams)
}

var DefaultShowResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.GetController().Resource.SelectById(c.GetApplication().DB, c.Param("id")).
		ApplyQueryParams(params)
}

var DefaultCreateResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	cm, err := c.GetCreateModel()
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewInsertQuery(c.GetApplication().DB, cm).
		ApplyQueryParams(params)
}

var DefaultUpdateResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	um, err := c.GetUpdateModel()
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewUpdateQuery(c.GetApplication().DB, um).
		ApplyQueryParams(params)
}

var DefaultDeleteResourceHandler HandlerFunc = func(c *Context) margo.Response {
	return c.GetController().Resource.DeleteById(c.GetApplication().DB, c.Param("id"))
}
