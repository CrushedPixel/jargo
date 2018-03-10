package jargo

type CreatePayloadHandlerFunc func(request *CreateRequest, payload interface{}) Response

type BeforeCreateQueryHandlerFunc func(request *CreateRequest, query *Query) *Query

type CreateResultHandlerFunc func(request *CreateRequest, result interface{}) Response

// CreateAction is a customizable CreateHandler.
//
// By default, it supports Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
type CreateAction struct {
	payloadHandler CreatePayloadHandlerFunc
	beforeQuery    BeforeCreateQueryHandlerFunc
	resultHandler  CreateResultHandlerFunc
}

// NewCreateAction creates a new default CreateAction instance.
func NewCreateAction() *CreateAction {
	return &CreateAction{}
}

func (a *CreateAction) Handle(req *CreateRequest) Response {
	// parse create payload
	instance, err := req.Resource().ParseJsonapiPayload(req.Payload(), req.Application().Validate())
	if err != nil {
		return NewErrorResponse(err)
	}

	// if set, apply payload handler
	if a.payloadHandler != nil {
		if res := a.payloadHandler(req, instance); res != nil {
			return res
		}
	}

	// create insert query
	q := req.Resource().InsertInstance(req.DB(), instance).
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

// BeforeQueryHandlerFunc sets the BeforeCreateQueryHandlerFunc
// to be applied before executing the query,
// replacing the existing handler function.
func (a *CreateAction) BeforeQueryHandlerFunc(f BeforeCreateQueryHandlerFunc) {
	a.beforeQuery = f
}

// ResultHandler sets the CreateResultHandlerFunc to be
// used, replacing the existing handler function.
func (a *CreateAction) ResultHandlerFunc(f CreateResultHandlerFunc) {
	a.resultHandler = f
}