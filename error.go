package jargo

import (
	"github.com/google/jsonapi"
	"strconv"
	"github.com/satori/go.uuid"
	"strings"
	"net/http"
)

const (
	invalidQueryParams = "INVALID_QUERY_PARAMS"
	invalidFilter      = "INVALID_FILTER"
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

func InvalidQueryParamsError() *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, invalidQueryParams)
}

func InvalidFilterError(detail string) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, invalidFilter, detail)
}

