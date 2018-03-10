package jargo

type UpdatePayloadHandlerFunc func(request *UpdateRequest, payload interface{}) Response

type BeforeUpdateQueryHandlerFunc func(request *UpdateRequest, query *Query) *Query

type UpdateResultHandlerFunc func(request *UpdateRequest, result interface{}) Response

// UpdateAction is a customizable UpdateHandler.
//
// By default, it supports Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#crud-updating
type UpdateAction struct {
	payloadHandler UpdatePayloadHandlerFunc
	beforeQuery    BeforeUpdateQueryHandlerFunc
	resultHandler  UpdateResultHandlerFunc
}

// NewUpdateAction creates a new default UpdateAction instance.
func NewUpdateAction() *UpdateAction {
	return &UpdateAction{}
}

func (a *UpdateAction) Handle(req *UpdateRequest) Response {
	// fetch existing resource from database
	existing, err := req.Resource().SelectById(req.DB(), req.ResourceId()).Result()
	if err != nil {
		return NewErrorResponse(err)
	}
	if existing == nil {
		return NewErrorResponse(err)
	}

	// parse update payload, applying it to existing instance
	instance, err := req.Resource().ParseJsonapiUpdatePayload(req.Payload(), existing, req.Application().Validate())

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
