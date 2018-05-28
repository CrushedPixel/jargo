package jargo

// CreateRequestHandlerFunc allows handling of the request object
// before any other action is taken.
// If a Response is returned, it is sent to the client,
// and no further action is taken.
type CreateRequestHandlerFunc func(request *CreateRequest) Response

// CreatePayloadHandlerFunc allows control over the resource instance the user is creating.
// The instance parsed from the request payload is replaced with the instance
// returned by the handler.
// If a Response is returned, it is immediately sent to the client,
// and the resource instance is not created.
type CreatePayloadHandlerFunc func(request *CreateRequest, payload interface{}) (interface{}, Response)

type BeforeCreateQueryHandlerFunc func(request *CreateRequest, query *Query) *Query

type CreateResultHandlerFunc func(request *CreateRequest, result interface{}) Response

// CreateAction is a customizable CreateHandler.
//
// By default, it supports Sparse Fieldsets
// according to the JSON API spec.
// http://jsonapi.org/format/#crud-creating
type CreateAction struct {
	requestHandler CreateRequestHandlerFunc
	payloadHandler CreatePayloadHandlerFunc
	beforeQuery    BeforeCreateQueryHandlerFunc
	resultHandler  CreateResultHandlerFunc
}

// NewCreateAction creates a new default CreateAction instance.
func NewCreateAction() *CreateAction {
	return &CreateAction{}
}

func (a *CreateAction) Handle(req *CreateRequest) Response {
	// if set, apply request handler
	if a.requestHandler != nil {
		if res := a.requestHandler(req); res != nil {
			return res
		}
	}

	// parse create payload
	instance, err := req.Resource().ParseJsonapiPayload(req.Payload(), req.Application().Validate(), true)
	if err != nil {
		return NewErrorResponse(err)
	}

	// if set, apply payload handler
	if a.payloadHandler != nil {
		var res Response
		if instance, res = a.payloadHandler(req, instance); res != nil {
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

// CreateRequestHandlerFunc sets the CreateRequestHandlerFunc
// to be applied, replacing the existing handler function.
func (a *CreateAction) CreateRequestHandlerFunc(f CreateRequestHandlerFunc) {
	a.requestHandler = f
}

// CreatePayloadHandlerFunc sets the CreatePayloadHandlerFunc
// to be applied after parsing the resource instance
// created by the user, replacing the existing handler function.
func (a *CreateAction) CreatePayloadHandlerFunc(f CreatePayloadHandlerFunc) {
	a.payloadHandler = f
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
