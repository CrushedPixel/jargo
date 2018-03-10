package jargo

type BeforeQueryHandlerFunc func(*Query) *Query

type IndexResultHandlerFunc func(request *IndexRequest, result interface{}) Response

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

// NewIndexAction creates a new default IndexAction instance.
func NewIndexAction() *IndexAction {
	return &IndexAction{}
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

	// if set, apply result handler
	if a.resultHandler != nil {
		if res := a.resultHandler(req, result); res != nil {
			return res
		}
	}

	// default result handling
	return req.Resource().Response(result, req.Fields())
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
