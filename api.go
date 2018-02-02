package jargo

func ParseCreatePayload(c *Context, request Request) (interface{}, error) {
	return c.Resource().ParseJsonapiPayloadString(request.Payload(), c.Application().Validate())
}

func ParseUpdatePayload(c *Context, request Request) (interface{}, error) {
	instance, err := c.Resource().SelectById(c.DB(), request.ResourceId()).Result()
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrNotFound
	}

	return c.Resource().ParseJsonapiUpdatePayloadString(request.Payload(), instance, c.Application().Validate())
}

// DefaultIndexResourceHandler is the HandlerFunc
// used by the builtin JSON API Index Action.
// It supports Pagination, Sorting, Filtering and Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultIndexResourceHandler(c *Context, r Request) Response {
	return c.Resource().Select(c.DB()).
		Filters(r.Filters()).
		Fields(r.FieldSet()).
		Sort(r.SortFields()).
		Pagination(r.Pagination())
}

// DefaultShowResourceHandler is the HandlerFunc
// used by the builtin JSON API Show Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultShowResourceHandler(c *Context, r Request) Response {
	return c.Resource().SelectById(c.DB(), r.ResourceId()).
		Fields(r.FieldSet())
}

// DefaultCreateResourceHandler is the HandlerFunc
// used by the builtin JSON API Create Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
func DefaultCreateResourceHandler(c *Context, r Request) Response {
	m, err := ParseCreatePayload(c, r)
	if err != nil {
		return ErrorResponse(err)
	}

	return c.Resource().InsertInstance(c.DB(), m).
		Fields(r.FieldSet())
}

// DefaultUpdateResourceHandler is the HandlerFunc
// used by the builtin JSON API Update Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
func DefaultUpdateResourceHandler(c *Context, r Request) Response {
	m, err := ParseUpdatePayload(c, r)
	if err != nil {
		return ErrorResponse(err)
	}

	return c.Resource().UpdateInstance(c.DB(), m).
		Fields(r.FieldSet())
}

// DefaultDeleteResourceHandler is the HandlerFunc
// used by the builtin JSON API Delete Action.
// http://jsonapi.org/format/#crud-deleting
func DefaultDeleteResourceHandler(c *Context, r Request) Response {
	return c.Resource().DeleteById(c.DB(), r.ResourceId())
}
