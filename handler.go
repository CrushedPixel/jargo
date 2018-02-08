package jargo

import "github.com/crushedpixel/ferry"

// A MiddlewareFunc is a function that can be executed for any request type.
type MiddlewareFunc func(context *Context) Response

// An IndexHandlerFunc is a function handling an index request.
type IndexHandlerFunc func(context *Context, request *IndexRequest) Response
type indexHandlerChain []IndexHandlerFunc

// A ShowHandlerFunc is a function handling a show request.
type ShowHandlerFunc func(context *Context, request *ShowRequest) Response
type showHandlerChain []ShowHandlerFunc

// A CreateHandlerFunc is a function handling a create request.
type CreateHandlerFunc func(context *Context, request *CreateRequest) Response
type createHandlerChain []CreateHandlerFunc

// An UpdateHandlerFunc is a function handling an update request.
type UpdateHandlerFunc func(context *Context, request *UpdateRequest) Response
type updateHandlerChain []UpdateHandlerFunc

// A DeleteHandlerFunc is a function handling a delete request.
type DeleteHandlerFunc func(context *Context, request *DeleteRequest) Response
type deleteHandlerChain []DeleteHandlerFunc

func (c indexHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) *ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return responseToFerry(res)
			}
		}

		// create IndexRequest instance from request
		req, err := parseIndexRequest(context)
		if err != nil {
			return responseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return responseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c showHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) *ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return responseToFerry(res)
			}
		}

		// create ShowRequest instance from request
		req, err := parseShowRequest(context)
		if err != nil {
			return responseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return responseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c createHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) *ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return responseToFerry(res)
			}
		}

		// create CreateRequest instance from request
		req, err := parseCreateRequest(context)
		if err != nil {
			return responseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return responseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c updateHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) *ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return responseToFerry(res)
			}
		}

		// create UpdateRequest instance from request
		req, err := parseUpdateRequest(context)
		if err != nil {
			return responseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return responseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c deleteHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) *ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return responseToFerry(res)
			}
		}

		// create DeleteRequest instance from request
		req, err := parseDeleteRequest(context)
		if err != nil {
			return responseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return responseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func responseToFerry(res Response) *ferry.Response {
	payload, err := res.Payload()
	if err != nil {
		res = NewErrorResponse(err)
		payload, err = res.Payload()
		if err != nil {
			panic(err)
		}
	}

	return &ferry.Response{
		Status:  res.Status(),
		Payload: payload,
	}
}
