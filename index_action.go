package jargo

type BeforeQueryHandlerFunc func(*Query) *Query

type IndexResultHandlerFunc func(*IndexRequest, interface{}) Response

var DefaultIndexResultHandlerFunc = func(req *IndexRequest, result interface{}) Response {
	return req.Resource().Response(result, req.Fields())
}

// IndexAction is a customizable IndexHandler.
//
// By default, it supports Pagination, Sorting,
// Filtering and Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#fetching
type IndexAction struct {
	beforeQuery   BeforeQueryHandlerFunc
	resultHandler IndexResultHandlerFunc
}

// NewIndexAction creates a new default index action.
func NewIndexAction() *IndexAction {
	return &IndexAction{
		resultHandler: DefaultIndexResultHandlerFunc,
	}
}

func (a *IndexAction) Handle(req *IndexRequest) Response {
	// create index query
	q := req.Resource().Select(req.DB()).
		Filters(req.Filters()).
		Fields(req.Fields()).
		Pagination(req.Pagination())

	// if set, apply beforeQuery handler
	if a.beforeQuery != nil {
		q = a.beforeQuery(q)
	}

	// execute query
	result, err := q.Result()
	if err != nil {
		return NewErrorResponse(err)
	}

	return a.resultHandler(req, result)
}

// BeforeQueryHandlerFunc sets the BeforeQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *IndexAction) BeforeQueryHandlerFunc(f BeforeQueryHandlerFunc) {
	a.beforeQuery = f
}

// ResultHandler sets the IndexResultHandlerFunc to be
// used, replacing the existing handler function.
func (a *IndexAction) ResultHandlerFunc(f IndexResultHandlerFunc) {
	a.resultHandler = f
}
