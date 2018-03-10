package jargo

func ParseUpdatePayload(req *UpdateRequest) (interface{}, error) {
	instance, err := req.Resource().SelectById(req.DB(), req.ResourceId()).Result()
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrNotFound
	}

	return req.Resource().ParseJsonapiUpdatePayload(req.Payload(), instance, req.Application().Validate())
}

// DefaultUpdateResourceHandler is the HandlerFunc
// used by the builtin JSON API Update Action.
// It supports Sparse Fieldsets according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
var DefaultUpdateResourceHandler = UpdateHandlerFunc(func(req *UpdateRequest) Response {
	m, err := ParseUpdatePayload(req)
	if err != nil {
		return NewErrorResponse(err)
	}

	return req.Resource().UpdateInstance(req.DB(), m).
		Fields(req.Fields())
})

// DefaultDeleteResourceHandler is the HandlerFunc
// used by the builtin JSON API Delete Action.
// http://jsonapi.org/format/#crud-deleting
var DefaultDeleteResourceHandler = DeleteHandlerFunc(func(req *DeleteRequest) Response {
	return req.Resource().DeleteById(req.DB(), req.ResourceId())
})
