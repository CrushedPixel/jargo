package jargo

type BeforeShowQueryHandlerFunc func(request *ShowRequest, query *Query) *Query

type ShowResultHandlerFunc func(request *ShowRequest, result interface{}) Response

// ShowAction is a customizable ShowHandler.
//
// By default, it supports Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
type ShowAction struct {
	beforeQuery   BeforeShowQueryHandlerFunc
	resultHandler ShowResultHandlerFunc
}

// NewShowAction creates a new default ShowAction instance.
func NewShowAction() *ShowAction {
	return &ShowAction{}
}

func (a *ShowAction) Handle(req *ShowRequest) Response {
	// create show query
	q := req.Resource().SelectById(req.DB(), req.ResourceId()).
		Fields(req.Fields())

	// if set, apply beforeQuery handler
	if a.beforeQuery != nil {
		q = a.beforeQuery(req, q)
	}

	// execute query
	result, err := q.Result()
	if err != nil {
		return NewErrorResponse(err)
	}

	// if set, apply result handler
	if a.resultHandler != nil {
		if res := a.resultHandler(req, result); res != nil {
			return res
		}
	}

	// default result handling
	if result == nil {
		return ErrNotFound
	}
	return req.Resource().Response(result, req.Fields())
}

// BeforeQueryHandlerFunc sets the BeforeShowQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *ShowAction) BeforeQueryHandlerFunc(f BeforeShowQueryHandlerFunc) {
	a.beforeQuery = f
}

// ResultHandler sets the ShowResultHandlerFunc to be
// used, replacing the existing handler function.
func (a *ShowAction) ResultHandlerFunc(f ShowResultHandlerFunc) {
	a.resultHandler = f
}
