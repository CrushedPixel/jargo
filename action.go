package jargo

import (
	"github.com/crushedpixel/margo"
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"net/http"
	"strings"
)

type Action struct {
	handlers []HandlerFunc
}

// NewAction returns a new Action
// from one or more HandlerFuncs.
//
// Panics if no handler is provided.
func NewAction(handlers ...HandlerFunc) *Action {
	return &Action{
		handlers: handlers,
	}
}

// NewJSONAPIAction returns a new Action,
// inserting middleware validating requests according to
// the JSON API spec before other handlers.
// http://jsonapi.org/format/#content-negotiation-clients
//
// Panics if no handler is provided.
func NewJSONAPIAction(handlers ...HandlerFunc) *Action {
	// prepend contentTypeMiddleware
	handlers = append([]HandlerFunc{contentTypeMiddleware}, handlers...)
	return NewAction(handlers...)
}

// Handlers returns a HandlerChain to be called when
// an Application handles a request for this Action.
func (a *Action) Handlers(app *Application) HandlerChain {
	// prepend injectApplicationMiddleware
	handlers := append([]HandlerFunc{injectApplicationMiddleware(app)}, a.handlers...)
	return HandlerChain(handlers)
}

// A HandlerFunc is a function handling a request.
type HandlerFunc func(context *Context) margo.Response

// A HandlerChain is a slice of handler functions to be executed in order.
// If a HandlerFunc returns a Response value, the Response is sent to the client,
// otherwise, the next HandlerFunc in the chain is executed.
// The last HandlerFunc in the chain is expected to return a Response value.
type HandlerChain []HandlerFunc

// ToMargoHandler converts a HandlerChain into a single margo.HandlerFunc.
func (chain HandlerChain) ToMargoHandler() margo.HandlerFunc {
	return func(context *gin.Context) margo.Response {
		if len(chain) < 1 {
			return nil
		}

		// wrap gin context in a jargo context
		jargoContext := &Context{
			Context: context,
		}
		for _, h := range chain {
			if res := h(jargoContext); res != nil {
				return res
			}
		}
		return nil
	}
}

// DefaultIndexResourceHandler is the HandlerFunc
// used by the builtin JSON API index Action.
// It supports Pagination, Sorting, Filtering and Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultIndexResourceHandler(c *Context) margo.Response {
	filters, err := c.Filters()
	if err != nil {
		return NewErrorResponse(err)
	}

	fields, err := c.FieldSet()
	if err != nil {
		return NewErrorResponse(err)
	}

	sort, err := c.SortFields()
	if err != nil {
		return NewErrorResponse(err)
	}

	pagination, err := c.Pagination()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().Select(c.DB()).
		Filters(filters).
		Fields(fields).
		Sort(sort).
		Pagination(pagination)
}

// DefaultShowResourceHandler is the HandlerFunc
// used by the builtin JSON API show Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultShowResourceHandler(c *Context) margo.Response {
	fields, err := c.FieldSet()
	if err != nil {
		return NewErrorResponse(err)
	}

	id, err := c.ResourceId()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().SelectById(c.DB(), id).
		Fields(fields)
}

// DefaultCreateResourceHandler is the HandlerFunc
// used by the builtin JSON API create Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
func DefaultCreateResourceHandler(c *Context) margo.Response {
	fields, err := c.FieldSet()
	if err != nil {
		return NewErrorResponse(err)
	}

	m, err := c.CreateModel()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().InsertOne(c.DB(), m).
		Fields(fields)
}

// DefaultUpdateResourceHandler is the HandlerFunc
// used by the builtin JSON API update Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
func DefaultUpdateResourceHandler(c *Context) margo.Response {
	fields, err := c.FieldSet()
	if err != nil {
		return NewErrorResponse(err)
	}

	m, err := c.UpdateModel()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().UpdateOne(c.DB(), m).
		Fields(fields)
}

// DefaultDeleteResourceHandler is the HandlerFunc
// used by the builtin JSON API delete Action.
// http://jsonapi.org/format/#crud-deleting
func DefaultDeleteResourceHandler(c *Context) margo.Response {
	id, err := c.ResourceId()
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().DeleteById(c.DB(), id)
}

// injectApplicationMiddleware returns a HandlerFunc
// setting the Context's Application.
func injectApplicationMiddleware(app *Application) HandlerFunc {
	return func(c *Context) margo.Response {
		c.setApplication(app)
		return nil
	}
}

// contentTypeMiddleware is a HandlerFunc validating JSON API requests
// according to JSON API spec.
// http://jsonapi.org/format/#content-negotiation-clients
func contentTypeMiddleware(c *Context) margo.Response {
	// if Content-Type header not the jsonapi media type,
	// return 415 Unsupported Media Type
	ct := c.Request.Header.Get("Content-Type")
	if ct != jsonapi.MediaType &&
		c.Request.Method != http.MethodGet &&
		c.Request.Method != http.MethodDelete {
		return ErrUnsupportedMediaType
	}

	var contains, exact bool
	for _, accept := range c.Request.Header["Accept"] {
		if jsonapi.MediaType == accept {
			exact = true
			break
		}

		if strings.Contains(accept, jsonapi.MediaType) {
			contains = true
		}
	}

	// if accept header contains media type but never unmodified,
	// return 406 Not Acceptable
	if contains && !exact {
		return ErrNotAcceptable
	}

	return nil
}
