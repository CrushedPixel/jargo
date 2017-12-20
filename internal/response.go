package internal

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"errors"
	"github.com/google/jsonapi"
)

var ErrInvalidDataResponse = errors.New("expected slice of struct pointers or struct pointer as data value")

type resourceResponse struct {
	resource *Resource
	data     interface{}
	fieldSet *FieldSet
	status   int
}

// satisfies margo.Response
func (r *resourceResponse) Send(c *gin.Context) error {
	// create jsonapi response model for fieldSet
	jsonapiType := r.resource.jsonapiModel(r.fieldSet)

	var jsonapiData interface{} // the data to marshal

	val := reflect.ValueOf(r.data)
	if val.Kind() == reflect.Slice {
		if val.Type().Elem().Kind() != reflect.Ptr ||
			val.Type().Elem().Elem().Kind() != reflect.Struct {
			return ErrInvalidDataResponse
		}

		// data is slice of struct pointers
		dataSlice := make([]interface{}, 0)

		for i := 0; i < val.Len(); i++ {
			entry := val.Index(i)
			data := reflect.New(jsonapiType)

			r.fieldSet.applyValues(&entry, &data)
			dataSlice = append(dataSlice, data.Interface())
		}

		jsonapiData = dataSlice
	} else if val.Kind() == reflect.Ptr {
		if val.Type().Elem().Kind() != reflect.Struct {
			return ErrInvalidDataResponse
		}

		// data is struct pointer
		data := reflect.New(jsonapiType)
		r.fieldSet.applyValues(&val, &data)

		jsonapiData = data.Interface()
	} else {
		return ErrInvalidDataResponse
	}

	c.Status(r.status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalPayloadWithoutIncluded(c.Writer, jsonapiData)
}
