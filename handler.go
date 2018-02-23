package jargo

import "github.com/crushedpixel/ferry"

// A HandlerFunc is a function handling a generic jargo request.
type HandlerFunc func(context *Context) Response
type handlerChain []HandlerFunc

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

func (c handlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware and handlers
		for _, m := range append(cont.middleware, c...) {
			res := m(context)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c indexHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create IndexRequest instance from request
		req, err := ParseIndexRequest(context)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c showHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create ShowRequest instance from request
		req, err := ParseShowRequest(context)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c createHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create CreateRequest instance from request
		req, err := ParseCreateRequest(context)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c updateHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create UpdateRequest instance from request
		req, err := ParseUpdateRequest(context)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c deleteHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(ctx *ferry.Context) ferry.Response {
		context := &Context{
			Context:     ctx,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(context)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create DeleteRequest instance from request
		req, err := ParseDeleteRequest(context)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(context, req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

// ResponseToFerry creates a ferry.Response from a Response,
// invoking its Payload() method and handling any errors.
func ResponseToFerry(res Response) ferry.Response {
	payload, err := res.Payload()
	if err != nil {
		res = NewErrorResponse(err)
		payload, err = res.Payload()
		if err != nil {
			panic(err)
		}
	}

	return ferry.NewResponse(res.Status(), payload)
}
