package jargo

// UpdateRequestHandlerFunc allows handling of the request object
// before any other action is taken.
// If a Response is returned, it is sent to the client,
// and no further action is taken.
type UpdateRequestHandlerFunc func(request *UpdateRequest) Response

type UpdatePayloadHandlerFunc func(request *UpdateRequest, payload interface{}) Response

type BeforeUpdateQueryHandlerFunc func(request *UpdateRequest, query *Query) *Query

type UpdateResultHandlerFunc func(request *UpdateRequest, result interface{}) Response

// UpdateAction is a customizable UpdateHandler.
//
// By default, it supports Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
type UpdateAction struct {
	requestHandler UpdateRequestHandlerFunc
	payloadHandler UpdatePayloadHandlerFunc
	beforeQuery    BeforeUpdateQueryHandlerFunc
	resultHandler  UpdateResultHandlerFunc
}

// NewUpdateAction creates a new default UpdateAction instance.
func NewUpdateAction() *UpdateAction {
	return &UpdateAction{}
}

func (a *UpdateAction) Handle(req *UpdateRequest) Response {
	// if set, apply request handler
	if a.requestHandler != nil {
		if res := a.requestHandler(req); res != nil {
			return res
		}
	}

	// fetch existing resource from database
	existing, err := req.Resource().SelectById(req.DB(), req.ResourceId()).Result()
	if err != nil {
		return NewErrorResponse(err)
	}
	if existing == nil {
		return NewErrorResponse(err)
	}

	// parse update payload, applying it to existing instance
	instance, err := req.Resource().ParseJsonapiUpdatePayload(req.Payload(), existing,
		req.Application().Validate(), true)
	if err != nil {
		return NewErrorResponse(err)
	}

	// if set, apply payload handler
	if a.payloadHandler != nil {
		if res := a.payloadHandler(req, instance); res != nil {
			return res
		}
	}

	// create update query
	q := req.Resource().UpdateInstance(req.DB(), instance).
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
	return req.Resource().Response(result, req.Fields())
}

// UpdateRequestHandlerFunc sets the UpdateRequestHandlerFunc
// to be applied, replacing the existing handler function.
func (a *UpdateAction) UpdateRequestHandlerFunc(f UpdateRequestHandlerFunc) {
	a.requestHandler = f
}

// BeforeQueryHandlerFunc sets the BeforeUpdateQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *UpdateAction) BeforeQueryHandlerFunc(f BeforeUpdateQueryHandlerFunc) {
	a.beforeQuery = f
}

// ResultHandler sets the UpdateResultHandlerFunc to be
// used, replacing the existing handler function.
func (a *UpdateAction) ResultHandlerFunc(f UpdateResultHandlerFunc) {
	a.resultHandler = f
}
