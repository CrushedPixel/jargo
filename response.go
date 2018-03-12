package jargo

import (
	"errors"
	"github.com/google/jsonapi"
	"github.com/json-iterator/go"
)

type Response interface {
	// Status returns the HTTP Status
	// for the response.
	Status() int
	// Payload returns the JSON API payload
	// to send to the client.
	Payload() (string, error)
}

type response struct {
	status  int
	payload string
}

func (r *response) Status() int {
	return r.status
}

func (r *response) Payload() (string, error) {
	return r.payload, nil
}

func NewResponse(status int, payload string) Response {
	return &response{
		status:  status,
		payload: payload,
	}
}

var errDataNil = errors.New("resource response data is nil")

type resourceResponse struct {
	data       interface{} // jsonapi model data
	collection bool

	fieldSet *FieldSet
	status   int
}

func (r *resourceResponse) Status() int {
	return r.status
}

func (r *resourceResponse) Payload() (string, error) {
	if r.data == nil {
		return "", errDataNil
	}

	p, err := jsonapi.Marshal(r.data)
	if err != nil {
		return "", err
	}

	var bytes []byte
	if r.collection {
		payload := p.(*jsonapi.ManyPayload)
		payload.Included = nil
		for _, node := range payload.Data {
			r.fieldSet.applyToJsonapiNode(node)
		}

		bytes, err = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(payload)
		if err != nil {
			return "", err
		}
	} else {
		payload := p.(*jsonapi.OnePayload)
		payload.Included = nil
		r.fieldSet.applyToJsonapiNode(payload.Data)

		bytes, err = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(payload)
		if err != nil {
			return "", err
		}
	}

	return string(bytes), nil
}
