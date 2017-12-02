package jargo

import (
	"github.com/google/jsonapi"
	"strconv"
	"github.com/satori/go.uuid"
	"strings"
	"net/http"
)

const (
	codeInvalidQueryParams = "INVALID_QUERY_PARAMS"
)

func NewErrorObject(status int, code string, detail ...string) *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(status),
		Code:   code,
		Detail: strings.Join(detail, ", "),
	}
}

func ToErrorResponse(e *jsonapi.ErrorObject) *ErrorResponse {
	status, err := strconv.Atoi(e.Status)
	if err != nil {
		panic(err)
	}
	return NewErrorResponse(status, e)
}

func invalidQueryParams(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, codeInvalidQueryParams, err.Error())
}
