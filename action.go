package jargo

import (
	"crushedpixel.net/margo"
	"fmt"
	"net/http"
)

type Route struct {
	Method string
	Path   string
}

type HandlerFunc func(context *Context) margo.Response

type Action struct {
	Enabled  bool
	handlers []HandlerFunc
}

func NewAction(handlers ...HandlerFunc) *Action {
	return &Action{
		Enabled:  true,
		handlers: handlers,
	}
}

type Actions map[Route]*Action

var showRoute = Route{Method: http.MethodGet, Path: "/:id"}
var indexRoute = Route{Method: http.MethodGet, Path: "/"}
var createRoute = Route{Method: http.MethodPost, Path: "/"}
var updateRoute = Route{Method: http.MethodPatch, Path: "/:id"}
var deleteRoute = Route{Method: http.MethodDelete, Path: "/:id"}

func (a Actions) GetShowAction() *Action {
	return a[showRoute]
}

func (a Actions) SetShowAction(action *Action) {
	a[showRoute] = action
}

func (a Actions) GetIndexAction() *Action {
	return a[indexRoute]
}

func (a Actions) SetIndexAction(action *Action) {
	a[indexRoute] = action
}

func (a Actions) GetCreateAction() *Action {
	return a[createRoute]
}

func (a Actions) SetCreateAction(action *Action) {
	a[createRoute] = action
}

func (a Actions) GetUpdateAction() *Action {
	return a[updateRoute]
}

func (a Actions) SetUpdateAction(action *Action) {
	a[updateRoute] = action
}

func (a Actions) GetDeleteAction() *Action {
	return a[deleteRoute]
}

func (a Actions) SetDeleteAction(action *Action) {
	a[deleteRoute] = action
}

func (a *Action) toEndpoint(c *Controller, route Route) *margo.Endpoint {
	fullPath := fmt.Sprintf("%s%s", c.Resource.Name, route.Path)

	endpoint := margo.NewEndpoint(route.Method, fullPath,
		toMargoHandler(
			injectControllerMiddleware(c),
			contentTypeMiddleware,
		),
		toMargoHandler(a.handlers...))

	return endpoint
}

func toMargoHandler(handlers ...HandlerFunc) margo.HandlerFunc {
	return func(c *margo.Context) margo.Response {
		context := &Context{c}

		for _, h := range handlers {
			if res := h(context); res != nil {
				return res
			}
		}

		return nil
	}
}

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

var IndexResourceHandler HandlerFunc = func(c *Context) margo.Response {
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

var ShowResourceQuery = func(c *Context) *Query {
	q := c.GetController().Resource.SelectOne(c.GetApplication().DB)
	q.Where("id = ?", c.Params.ByName("id"))

	return q
}

var ShowResourceHandler HandlerFunc = func(c *Context) margo.Response {
	params, err := c.GetQueryParams()
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewDataResponse(ShowResourceQuery(c), params.Fields)
}

var CreateResourceHandler HandlerFunc = func(c *Context) margo.Response {
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

var UpdateResourceHandler HandlerFunc = func(c *Context) margo.Response {
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
