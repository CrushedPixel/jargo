package jargo

import "github.com/crushedpixel/ferry"

// A HandlerFunc is a function handling a generic jargo request.
type HandlerFunc func(req *Request) Response
type handlerChain []HandlerFunc

// An IndexHandlerFunc is a function handling an index request.
type IndexHandlerFunc func(req *IndexRequest) Response
type indexHandlerChain []IndexHandlerFunc

// A ShowHandlerFunc is a function handling a show request.
type ShowHandlerFunc func(req *ShowRequest) Response
type showHandlerChain []ShowHandlerFunc

// A CreateHandlerFunc is a function handling a create request.
type CreateHandlerFunc func(req *CreateRequest) Response
type createHandlerChain []CreateHandlerFunc

// An UpdateHandlerFunc is a function handling an update request.
type UpdateHandlerFunc func(req *UpdateRequest) Response
type updateHandlerChain []UpdateHandlerFunc

// A DeleteHandlerFunc is a function handling a delete request.
type DeleteHandlerFunc func(req *DeleteRequest) Response
type deleteHandlerChain []DeleteHandlerFunc

func (c handlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		req := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware and handlers
		for _, m := range append(cont.middleware, c...) {
			res := m(req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c indexHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		base := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create IndexRequest instance from request
		req, err := ParseIndexRequest(base)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c showHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		base := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create ShowRequest instance from request
		req, err := ParseShowRequest(base)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c createHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		base := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create CreateRequest instance from request
		req, err := ParseCreateRequest(base)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c updateHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		base := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create UpdateRequest instance from request
		req, err := ParseUpdateRequest(base)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(req)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		panic("last handler in chain did not return a value")
	}
}

func (c deleteHandlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		base := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware
		for _, m := range cont.middleware {
			res := m(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create DeleteRequest instance from request
		req, err := ParseDeleteRequest(base)
		if err != nil {
			return NewErrorResponse(err)
		}

		// execute handlers
		for _, h := range c {
			res := h(req)
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
