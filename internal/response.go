package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"crushedpixel.net/jargo/api"
	"encoding/json"
)

type resourceResponse struct {
	data interface{} // resource model data

	resource api.Resource
	fieldSet api.FieldSet
	status   int
}

func (r *resourceResponse) Send(c *gin.Context) error {
	collection, err := r.resource.IsResourceModelCollection(r.data)
	if err != nil {
		return err
	}

	var bytes []byte

	if collection {
		instances, err := r.resource.ParseResourceModelCollection(r.data)
		if err != nil {
			return err
		}
		var jsonapiModels []interface{}
		for _, instance := range instances {
			jsonapiModelInstance, err := instance.ToJsonapiModel()
			if err != nil {
				return err
			}
			jsonapiModels = append(jsonapiModels, jsonapiModelInstance)
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
		instance, err := r.resource.ParseResourceModel(r.data)
		if err != nil {
			return err
		}
		jsonapiModelInstance, err := instance.ToJsonapiModel()
		if err != nil {
			return err
		}

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
	_, err = c.Writer.Write(bytes)
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
