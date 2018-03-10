package jargo

import "net/http"

type BeforeDeleteQueryHandlerFunc func(request *DeleteRequest, query *Query) *Query

// DeleteAction is a customizable DeleteHandler.
//
// It is implemented according to the JSON API spec.
// http://jsonapi.org/format/#crud-deleting
type DeleteAction struct {
	beforeQuery BeforeDeleteQueryHandlerFunc
}

// NewDeleteAction creates a new default DeleteAction instance.
func NewDeleteAction() *DeleteAction {
	return &DeleteAction{}
}

func (a *DeleteAction) Handle(req *DeleteRequest) Response {
	// create delete query
	q := req.Resource().DeleteById(req.DB(), req.ResourceId())

	// if set, apply beforeQuery handler
	if a.beforeQuery != nil {
		q = a.beforeQuery(req, q)
	}

	// execute query
	_, err := q.Result()
	if err != nil {
		return NewErrorResponse(err)
	}

	// return empty response
	return NewResponse(http.StatusNoContent, "")
}

// BeforeQueryHandlerFunc sets the BeforeDeleteQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *DeleteAction) BeforeQueryHandlerFunc(f BeforeDeleteQueryHandlerFunc) {
	a.beforeQuery = f
}
