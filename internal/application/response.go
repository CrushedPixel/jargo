package application

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"errors"
	"github.com/google/jsonapi"
	"net/http"
	"encoding/json"
	"fmt"
)

var ErrInvalidDataResponse = errors.New("expected slice of struct pointers or struct pointer as data value")
var ErrInvalidMarshalResult = errors.New("jsonapi.Marshal returned an unexpected value")

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
		println(fmt.Sprintf("Internal server error: %s", r.Error.Error())) // TODO use a proper logging library
		apiError = ApiErrInternalServerError
	}

	return apiError.Send(c)
}

type DataResponse struct {
	Data   interface{}
	Fields ResultFields
	Status int
}

func NewDataResponseAllFields(data interface{}) *DataResponse {
	return NewDataResponse(data, nil)
}

func NewDataResponse(data interface{}, fields ResultFields) *DataResponse {
	return NewDataResponseWithStatusCode(data, fields, http.StatusOK)
}

func NewDataResponseWithStatusCode(data interface{}, fields ResultFields, status int) *DataResponse {
	return &DataResponse{
		Data:   data,
		Fields: fields,
		Status: status,
	}
}

func (r *DataResponse) Send(c *gin.Context) error {
	val := reflect.ValueOf(r.Data)

	// data must be slice of struct pointers or struct pointer
	if !((val.Kind() == reflect.Slice &&
		val.Type().Elem().Kind() == reflect.Ptr &&
		val.Type().Elem().Elem().Kind() == reflect.Struct) ||
		(val.Kind() == reflect.Ptr &&
			val.Type().Elem().Kind() == reflect.Struct)) {
		return ErrInvalidDataResponse
	}

	payload, err := jsonapi.Marshal(val.Interface())
	if err != nil {
		return err
	}

	op, isOP := payload.(*jsonapi.OnePayload)
	if isOP {
		if r.Fields != nil {
			r.Fields.ApplyToNode(op.Data)
		}

		// clear included records, as they are not supported yet
		op.Included = []*jsonapi.Node{}
	}

	mp, isMP := payload.(*jsonapi.ManyPayload)
	if isMP {
		if r.Fields != nil {
			for _, node := range mp.Data {
				r.Fields.ApplyToNode(node)
			}
		}

		// clear included records, as they are not supported yet
		mp.Included = []*jsonapi.Node{}
	}

	if !isOP && !isMP {
		return ErrInvalidMarshalResult
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	c.Status(r.Status)
	c.Header("Content-Type", jsonapi.MediaType)
	_, err = c.Writer.Write(b)

	return err
}
