package jargo

// DefaultDeleteResourceHandler is the HandlerFunc
// used by the builtin JSON API Delete Action.
// http://jsonapi.org/format/#crud-deleting
var DefaultDeleteResourceHandler = DeleteHandlerFunc(func(req *DeleteRequest) Response {
	return req.Resource().DeleteById(req.DB(), req.ResourceId())
})
