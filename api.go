package jargo

func ParseCreatePayload(c *Context, req *CreateRequest) (interface{}, error) {
	return c.Resource().ParseJsonapiPayload(req.Payload, c.Application().Validate())
}

func ParseUpdatePayload(c *Context, req *UpdateRequest) (interface{}, error) {
	instance, err := c.Resource().SelectById(c.DB(), req.ResourceId).Result()
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrNotFound
	}

	return c.Resource().ParseJsonapiUpdatePayload(req.Payload, instance, c.Application().Validate())
}

// DefaultIndexResourceHandler is the HandlerFunc
// used by the builtin JSON API Index Action.
// It supports Pagination, Sorting, Filtering and Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultIndexResourceHandler(c *Context, req *IndexRequest) Response {
	return c.Resource().Select(c.DB()).
		Filters(req.Filters).
		Fields(req.Fields).
		Sort(req.SortFields).
		Pagination(req.Pagination)
}

// DefaultShowResourceHandler is the HandlerFunc
// used by the builtin JSON API Show Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#fetching
func DefaultShowResourceHandler(c *Context, req *ShowRequest) Response {
	return c.Resource().SelectById(c.DB(), req.ResourceId).
		Fields(req.Fields)
}

// DefaultCreateResourceHandler is the HandlerFunc
// used by the builtin JSON API Create Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
func DefaultCreateResourceHandler(c *Context, req *CreateRequest) Response {
	m, err := ParseCreatePayload(c, req)
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().InsertInstance(c.DB(), m).
		Fields(req.Fields)
}

// DefaultUpdateResourceHandler is the HandlerFunc
// used by the builtin JSON API Update Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
func DefaultUpdateResourceHandler(c *Context, req *UpdateRequest) Response {
	m, err := ParseUpdatePayload(c, req)
	if err != nil {
		return NewErrorResponse(err)
	}

	return c.Resource().UpdateInstance(c.DB(), m).
		Fields(req.Fields)
}

// DefaultDeleteResourceHandler is the HandlerFunc
// used by the builtin JSON API Delete Action.
// http://jsonapi.org/format/#crud-deleting
func DefaultDeleteResourceHandler(c *Context, req *DeleteRequest) Response {
	return c.Resource().DeleteById(c.DB(), req.ResourceId)
}
