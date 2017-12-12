package jargo

import (
	"crushedpixel.net/margo"
	"fmt"
	"net/http"
	"crushedpixel.net/jargo/models"
)

type Route struct {
	Method string
	Path   string
}

type HandlerFunc func(context *Context) interface{}

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
	fullPath := fmt.Sprintf("%s%s", c.Model.Name, route.Path)

	endpoint := margo.NewEndpoint(route.Method, fullPath,
		toMargoHandler(
			injectControllerMiddleware(c),
			contentTypeMiddleware(a),
		),
		toMargoHandler(a.handlers...))

	return endpoint
}

func toMargoHandler(handlers ...HandlerFunc) margo.HandlerFunc {
	return func(c *margo.Context) margo.Response {
		context := &Context{c}

		for _, h := range handlers {
			if res := h(context); res != nil {
				return NewResponse(res)
			}
		}

		return nil
	}
}

var IndexResourceQuery = func(c *Context) *models.Query {
	fp := c.GetFetchParams()

	q := c.GetController().Model.Select(c.GetApplication().DB)
	fp.ApplyToQuery(q)

	return q
}

var CreateResourceHandler HandlerFunc = func(c *Context) interface{} {
	cm := c.GetCreateModel()

	_, err := c.GetApplication().DB.Model(cm).Insert()
	if err != nil {
		panic(err)
	}

	return cm
}
