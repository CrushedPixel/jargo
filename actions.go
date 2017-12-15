package jargo

import (
	"crushedpixel.net/margo"
	"net/http"
)

var IndexResourceQuery = func(c *Context) (*Query, error) {
	q := c.GetController().Resource.Select(c.GetApplication().DB)

	params, err := c.GetQueryParams()
	if err != nil {
		return nil, err
	}
	q.ApplyQueryParams(params)

	indexParams, err := c.GetIndexQueryParams()
	if err != nil {
		return nil, err
	}
	q.ApplyIndexQueryParams(indexParams)

	return q, nil
}

var ShowResourceQuery = func(c *Context) *Query {
	q := c.GetController().Resource.SelectOne(c.GetApplication().DB)
	q.Where("id = ?", c.Params.ByName("id"))

	return q
}

var DefaultIndexResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	query, err := IndexResourceQuery(c)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewDataResponse(query, params.Fields)
}

var DefaultShowResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewDataResponse(ShowResourceQuery(c), params.Fields)
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

	_, err = c.GetApplication().DB.Model(cm).Insert()
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewDataResponseWithStatusCode(cm, params.Fields, http.StatusCreated)
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

	_, err = c.GetApplication().DB.Model(um).Update()
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewDataResponse(um, params.Fields)
}
