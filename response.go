package jargo

import (
	"encoding/json"
	"errors"
	"github.com/crushedpixel/margo"
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
)

var errDataNil = errors.New("resource response data is nil")

type resourceResponse struct {
	data       interface{} // jsonapi model data
	collection bool

	fieldSet *FieldSet
	status   int
}

func newResourceResponse(jsonapiModelData interface{}, collection bool, fieldSet *FieldSet, status int) margo.Response {
	return &resourceResponse{
		data:       jsonapiModelData,
		collection: collection,
		fieldSet:   fieldSet,
		status:     status,
	}
}

func (r *resourceResponse) Send(c *gin.Context) error {
	if r.data == nil {
		return errDataNil
	}

	p, err := jsonapi.Marshal(r.data)
	if err != nil {
		return err
	}

	var bytes []byte
	if r.collection {
		payload := p.(*jsonapi.ManyPayload)
		payload.Included = nil
		for _, node := range payload.Data {
			r.fieldSet.applyToJsonapiNode(node)
		}

		bytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	} else {
		payload := p.(*jsonapi.OnePayload)
		payload.Included = nil
		r.fieldSet.applyToJsonapiNode(payload.Data)

		bytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	c.Status(r.status)
	c.Header("Content-Type", jsonapi.MediaType)
	_, err = c.Writer.Write(bytes)
	return err
}
