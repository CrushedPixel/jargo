package jargo

import (
	"crushedpixel.net/margo"
	"fmt"
)

type HandlerFunc func(context *Context) Response

type Action struct {
	method   string
	path     string
	handlers []HandlerFunc
}

type BodyParameters struct {
	Data interface{} `json:"data"`
}

func NewAction(method string, path string, handlers ...HandlerFunc) *Action {
	return &Action{
		method, path, handlers,
	}
}

func (a *Action) toEndpoint(c *Controller) *margo.Endpoint {
	path := fmt.Sprintf("%s%s", c.BasePath, a.path)

	endpoint := margo.NewEndpoint(a.method, path,
		toMargoHandler(
			injectControllerMiddleware(c),
			fetchParamsMiddleware(a),
			contentTypeMiddleware(a),
		),
		toMargoHandler(a.handlers...))

	return endpoint
}

func toMargoHandler(handlers ...HandlerFunc) margo.HandlerFunc {
	return func(c *margo.Context) margo.Response {
		context := &Context{c}

		for _, h := range handlers {
			if response := h(context); response != nil {
				return response
			}
		}

		return nil
	}
}
