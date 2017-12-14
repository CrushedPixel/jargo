package jargo

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"errors"
	"github.com/google/jsonapi"
	"net/http"
)

var ErrInvalidDataResponse = errors.New("expected slice of struct pointers or struct pointer as data value")

type ErrorResponse struct {
	Error error
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{Error: err}
}

func (r *ErrorResponse) Send(c *gin.Context) error {
	apiError, ok := r.Error.(*ApiError)

	if !ok {
		// if error is not an api error, return internal server error response
		apiError = InternalServerError
	}

	return apiError.Send(c)
}

type DataResponse struct {
	Data   interface{}
	Fields *ResultFields
	Status int
}

func NewDataResponseAllFields(data interface{}) *DataResponse {
	return NewDataResponse(data, nil)
}

func NewDataResponse(data interface{}, fields *ResultFields) *DataResponse {
	return NewDataResponseWithStatusCode(data, fields, http.StatusOK)
}

func NewDataResponseWithStatusCode(data interface{}, fields *ResultFields, status int) *DataResponse {
	return &DataResponse{
		Data:   data,
		Fields: fields,
		Status: status,
	}
}

func (r *DataResponse) Send(c *gin.Context) error {
	val := reflect.ValueOf(r.Data)

	// data must be slice of struct pointers or struct pointer
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return ErrInvalidDataResponse
	}

	var value = val.Interface()

	// resolve query result if value is a query
	query, ok := value.(*Query)
	if ok {
		var err error
		value, err = query.GetValue()
		if err != nil {
			return err
		}
	}

	// TODO if Fields != nil, filter fields

	c.Status(r.Status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalPayloadWithoutIncluded(c.Writer, value)
}
