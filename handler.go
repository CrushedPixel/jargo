package jargo

import "github.com/crushedpixel/ferry"

// Handler handles a generic jargo request.
type Handler interface {
	Handle(*Request) Response
}

// HandlerFunc handles a generic jargo request.
type HandlerFunc func(req *Request) Response

func (h HandlerFunc) Handle(req *Request) Response {
	return h(req)
}

type handlerChain []Handler

// IndexHandler handles an index request.
type IndexHandler interface {
	Handle(request *IndexRequest) Response
}

// IndexHandlerFunc handles an index request.
type IndexHandlerFunc func(req *IndexRequest) Response

func (h IndexHandlerFunc) Handle(req *IndexRequest) Response {
	return h(req)
}

type indexHandlerChain []IndexHandler

// ShowHandler handles a show request.
type ShowHandler interface {
	Handle(*ShowRequest) Response
}

// ShowHandlerFunc handles a show request.
type ShowHandlerFunc func(req *ShowRequest) Response

func (h ShowHandlerFunc) Handle(req *ShowRequest) Response {
	return h(req)
}

type showHandlerChain []ShowHandler

// CreateHandler handles a create request.
type CreateHandler interface {
	Handle(*CreateRequest) Response
}

// CreateHandlerFunc handles a create request.
type CreateHandlerFunc func(req *CreateRequest) Response

func (h CreateHandlerFunc) Handle(req *CreateRequest) Response {
	return h(req)
}

type createHandlerChain []CreateHandler

// UpdateHandler handles an update request.
type UpdateHandler interface {
	Handle(*UpdateRequest) Response
}

// UpdateHandlerFunc handles an update request.
type UpdateHandlerFunc func(req *UpdateRequest) Response

func (h UpdateHandlerFunc) Handle(req *UpdateRequest) Response {
	return h(req)
}

type updateHandlerChain []UpdateHandler

// DeleteHandler handles a delete request.
type DeleteHandler interface {
	Handle(*DeleteRequest) Response
}

// DeleteHandlerFunc handles a delete request.
type DeleteHandlerFunc func(req *DeleteRequest) Response

func (h DeleteHandlerFunc) Handle(req *DeleteRequest) Response {
	return h(req)
}

type deleteHandlerChain []DeleteHandler

func (c handlerChain) toFerry(app *Application, cont *Controller) ferry.HandlerFunc {
	return func(r *ferry.Request) ferry.Response {
		req := &Request{
			Request:     r,
			application: app,
			resource:    cont.resource,
		}

		// execute middleware and handlers
		for _, m := range append(cont.middleware, c...) {
			res := m.Handle(req)
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
			res := m.Handle(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create IndexRequest instance from request
		req, err := ParseIndexRequest(base)
		if err != nil {
			return ResponseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h.Handle(req)
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
			res := m.Handle(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create ShowRequest instance from request
		req, err := ParseShowRequest(base)
		if err != nil {
			return ResponseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h.Handle(req)
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
			res := m.Handle(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create CreateRequest instance from request
		req, err := ParseCreateRequest(base)
		if err != nil {
			return ResponseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h.Handle(req)
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
			res := m.Handle(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create UpdateRequest instance from request
		req, err := ParseUpdateRequest(base)
		if err != nil {
			return ResponseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h.Handle(req)
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
			res := m.Handle(base)
			if res != nil {
				return ResponseToFerry(res)
			}
		}

		// create DeleteRequest instance from request
		req, err := ParseDeleteRequest(base)
		if err != nil {
			return ResponseToFerry(NewErrorResponse(err))
		}

		// execute handlers
		for _, h := range c {
			res := h.Handle(req)
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
