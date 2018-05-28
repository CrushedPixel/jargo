package jargo

// IndexRequestHandlerFunc allows handling of the request object
// before any other action is taken.
// If a Response is returned, it is sent to the client,
// and no further action is taken.
type IndexRequestHandlerFunc func(request *IndexRequest) Response

type BeforeIndexQueryHandlerFunc func(request *IndexRequest, query *Query) *Query

type IndexResultHandlerFunc func(request *IndexRequest, result interface{}) Response

// IndexAction is a customizable IndexHandler.
//
// By default, it supports Pagination, Sorting,
// Filtering and Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
type IndexAction struct {
	requestHandler IndexRequestHandlerFunc
	beforeQuery    BeforeIndexQueryHandlerFunc
	resultHandler  IndexResultHandlerFunc
}

// NewIndexAction creates a new default IndexAction instance.
func NewIndexAction() *IndexAction {
	return &IndexAction{}
}

func (a *IndexAction) Handle(req *IndexRequest) Response {
	// if set, apply request handler
	if a.requestHandler != nil {
		if res := a.requestHandler(req); res != nil {
			return res
		}
	}

	// create index query
	q := req.Resource().Select(req.DB()).
		Filters(req.Filters()).
		Fields(req.Fields()).
		Pagination(req.Pagination())

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
	return req.Resource().Response(result, req.Fields())
}

// IndexRequestHandlerFunc sets the IndexRequestHandlerFunc
// to be applied, replacing the existing handler function.
func (a *IndexAction) IndexRequestHandlerFunc(f IndexRequestHandlerFunc) {
	a.requestHandler = f
}

// BeforeQueryHandlerFunc sets the BeforeIndexQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *IndexAction) BeforeQueryHandlerFunc(f BeforeIndexQueryHandlerFunc) {
	a.beforeQuery = f
}

// ResultHandler sets the IndexResultHandlerFunc to be
// used, replacing the existing handler function.
func (a *IndexAction) ResultHandlerFunc(f IndexResultHandlerFunc) {
	a.resultHandler = f
}
