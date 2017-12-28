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
	JsonapiMiddleware bool
	handlers          []HandlerFunc
}

func NewAction(handlers ...HandlerFunc) *Action {
	return &Action{
		JsonapiMiddleware: false,
		handlers:          handlers,
	}
}

func NewJsonapiAction(handlers ...HandlerFunc) *Action {
	return &Action{
		JsonapiMiddleware: true,
		handlers:          handlers,
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

func (a Actions) SetAction(route *Route, action *Action) {
	a[*route] = action
}

func (a *Action) toEndpoint(c *Controller, route Route) *margo.Endpoint {
	namespace := c.Namespace
	// add trailing slash to namespace if missing
	if namespace != "" && namespace[len(namespace)-1] != '/' {
		namespace += "/"
	}

	middleware := []HandlerFunc{injectControllerMiddleware(c)}
	if a.JsonapiMiddleware {
		middleware = append(middleware, contentTypeMiddleware)
	}

	fullPath := fmt.Sprintf("%s%s%s", namespace, c.Resource.Name(), route.Path)
	endpoint := margo.NewEndpoint(route.Method, fullPath,
		toMargoHandler(c.Middleware...),
		toMargoHandler(middleware...),
		toMargoHandler(a.handlers...))

	return endpoint
}

func toMargoHandler(handlers ...HandlerFunc) margo.HandlerFunc {
	return func(c *margo.Context) margo.Response {
		context := &Context{c.Context}

		for _, h := range handlers {
			if res := h(context); res != nil {
				return res
			}
		}

		return nil
	}
}
