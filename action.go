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

	// TODO: array of allowed filter values
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
		contentTypeMiddleware(a),
		fetchParamsMiddleware(a),
		a.toMargoHandler(c))

	return endpoint
}

// converts the action's jargo handlers into a single margo handler
func (a *Action) toMargoHandler(cont *Controller) margo.HandlerFunc {
	return func(c *margo.Context) margo.Response {
		context := &Context{c}
		// inject controller into context
		context.Set(controller, cont)

		for _, h := range a.handlers {
			if response := h(context); response != nil {
				return response
			}
		}

		return nil
	}
}
