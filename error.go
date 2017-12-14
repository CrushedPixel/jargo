package jargo

import (
	"github.com/google/jsonapi"
	"net/http"
	"strconv"
	"github.com/satori/go.uuid"
	"github.com/gin-gonic/gin"
)

const (
	codeInvalidQueryParams  = "INVALID_QUERY_PARAMS"
	codeInvalidPayload      = "INVALID_PAYLOAD"
	codeForbidden           = "FORBIDDEN"
	codeInternalServerError = "INTERNAL_SERVER_ERROR"
)

var InternalServerError = NewApiError(
	http.StatusInternalServerError,
	codeInternalServerError,
	"internal server error",
)

type ApiError struct {
	Status int
	Code   string
	Detail string
}

// implements error
func (err *ApiError) Error() string {
	return err.Detail
}

// implements margo.Response
func (err *ApiError) Send(c *gin.Context) error {
	c.Status(err.Status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalErrors(c.Writer, []*jsonapi.ErrorObject{err.ToErrorObject()})
}

func (err *ApiError) ToErrorObject() *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(err.Status),
		Code:   err.Code,
		Detail: err.Detail,
	}
}

func NewApiError(status int, code string, detail string) *ApiError {
	return &ApiError{
		Status: status,
		Code:   code,
		Detail: detail,
	}
}

func invalidQueryParams(err error) *ApiError {
	return NewApiError(http.StatusBadRequest, codeInvalidQueryParams, err.Error())
}

func invalidPayload(err error) *ApiError {
	return NewApiError(http.StatusBadRequest, codeInvalidPayload, err.Error())
}

func forbidden(err error) *ApiError {
	return NewApiError(http.StatusForbidden, codeForbidden, err.Error())
}
