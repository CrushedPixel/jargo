package jargo

func ParseCreatePayload(c *Context) (interface{}, error) {
	return c.Resource().ParseJsonapiPayloadString(c.Payload(), c.Application().Validate())
}

func ParseUpdatePayload(c *Context) (interface{}, error) {
	instance, err := c.Resource().SelectById(c.DB(), c.ResourceId()).Result()
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrNotFound
	}

	return c.Resource().ParseJsonapiUpdatePayloadString(c.Payload(), instance, c.Application().Validate())
}

// DefaultIndexResourceHandler is the HandlerFunc
// used by the builtin JSON API Index Action.
// It supports Pagination, Sorting, Filtering and Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultIndexResourceHandler(c *Context) Response {
	return c.Resource().Select(c.DB()).
		Filters(c.Filters()).
		Fields(c.FieldSet()).
		Sort(c.SortFields()).
		Pagination(c.Pagination())
}

// DefaultShowResourceHandler is the HandlerFunc
// used by the builtin JSON API Show Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultShowResourceHandler(c *Context) Response {
	return c.Resource().SelectById(c.DB(), c.ResourceId()).
		Fields(c.FieldSet())
}

// DefaultCreateResourceHandler is the HandlerFunc
// used by the builtin JSON API Create Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
func DefaultCreateResourceHandler(c *Context) Response {
	m, err := ParseCreatePayload(c)
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().InsertInstance(c.DB(), m).
		Fields(c.FieldSet())
}

// DefaultUpdateResourceHandler is the HandlerFunc
// used by the builtin JSON API Update Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
func DefaultUpdateResourceHandler(c *Context) Response {
	m, err := ParseUpdatePayload(c)
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().UpdateInstance(c.DB(), m).
		Fields(c.FieldSet())
}

// DefaultDeleteResourceHandler is the HandlerFunc
// used by the builtin JSON API Delete Action.
// http://jsonapi.org/format/#crud-deleting
func DefaultDeleteResourceHandler(c *Context) Response {
	return c.Resource().DeleteById(c.DB(), c.ResourceId())
}
