package internal

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"github.com/google/jsonapi"
)

type resourceResponse struct {
	resource *Resource
	data     interface{}
	fieldSet *FieldSet
	status   int
}

// satisfies margo.Response
func (r *resourceResponse) Send(c *gin.Context) error {
	// create jsonapi response model for fieldSet
	jsonapiData, err := r.resource.registry.resourceModelToJsonapiModel(r.resource, reflect.ValueOf(r.data), r.resource.jsonapiModel(r.fieldSet))
	if err != nil {
		return err
	}

	c.Status(r.status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalPayloadWithoutIncluded(c.Writer, jsonapiData)
}
