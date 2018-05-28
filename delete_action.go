package jargo

import "net/http"

// DeleteRequestHandlerFunc allows handling of the request object
// before any other action is taken.
// If a Response is returned, it is sent to the client,
// and no further action is taken.
type DeleteRequestHandlerFunc func(request *DeleteRequest) Response

type BeforeDeleteQueryHandlerFunc func(request *DeleteRequest, query *Query) *Query

// DeleteAction is a customizable DeleteHandler.
//
// It is implemented according to the JSON API spec.
// http://jsonapi.org/format/#crud-deleting
type DeleteAction struct {
	requestHandler DeleteRequestHandlerFunc
	beforeQuery    BeforeDeleteQueryHandlerFunc
}

// NewDeleteAction creates a new default DeleteAction instance.
func NewDeleteAction() *DeleteAction {
	return &DeleteAction{}
}

func (a *DeleteAction) Handle(req *DeleteRequest) Response {
	// if set, apply request handler
	if a.requestHandler != nil {
		if res := a.requestHandler(req); res != nil {
			return res
		}
	}

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

// DeleteRequestHandlerFunc sets the DeleteRequestHandlerFunc
// to be applied, replacing the existing handler function.
func (a *DeleteAction) DeleteRequestHandlerFunc(f DeleteRequestHandlerFunc) {
	a.requestHandler = f
}

// BeforeQueryHandlerFunc sets the BeforeDeleteQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *DeleteAction) BeforeQueryHandlerFunc(f BeforeDeleteQueryHandlerFunc) {
	a.beforeQuery = f
}
