package internal

import (
	"encoding/json"
	"errors"
	"github.com/crushedpixel/jargo/api"
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
)

var errDataNil = errors.New("resource response data is nil")

type resourceResponse struct {
	data interface{} // resource model data

	resource api.Resource
	fieldSet api.FieldSet
	status   int
}

func (r *resourceResponse) Send(c *gin.Context) error {
	if r.data == nil {
		return errDataNil
	}
	collection := r.resource.IsResourceModelCollection(r.data)

	var bytes []byte
	if collection {
		instances := r.resource.ParseResourceModelCollection(r.data)

		var jsonapiModels []interface{}
		for _, instance := range instances {
			jsonapiModels = append(jsonapiModels, instance.ToJsonapiModel())
		}

		p, err := jsonapi.Marshal(jsonapiModels)
		if err != nil {
			return err
		}

		payload := p.(*jsonapi.ManyPayload)
		payload.Included = nil
		for _, node := range payload.Data {
			r.fieldSet.ApplyToJsonapiNode(node)
		}

		bytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	} else {
		instance := r.resource.ParseResourceModel(r.data)
		jsonapiModelInstance := instance.ToJsonapiModel()

		p, err := jsonapi.Marshal(jsonapiModelInstance)
		if err != nil {
			return err
		}

		payload := p.(*jsonapi.OnePayload)
		payload.Included = nil
		r.fieldSet.ApplyToJsonapiNode(payload.Data)

		bytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	c.Status(r.status)
	c.Header("Content-Type", jsonapi.MediaType)
	_, err := c.Writer.Write(bytes)
	return err
}

func newResourceResponse(resource api.Resource, data interface{}, fieldSet api.FieldSet, status int) *resourceResponse {
	return &resourceResponse{
		data:     data,
		resource: resource,
		fieldSet: fieldSet,
		status:   status,
	}
}
