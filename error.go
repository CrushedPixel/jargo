package jargo

import (
	"github.com/google/jsonapi"
	"strconv"
	"github.com/satori/go.uuid"
	"strings"
	"net/http"
	"crushedpixel.net/jargo/response"
)

const (
	codeInvalidQueryParams = "INVALID_QUERY_PARAMS"
	codeInvalidPayload     = "INVALID_PAYLOAD"
	codeForbidden          = "FORBIDDEN"
)

func NewErrorObject(status int, code string, detail ...string) *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(status),
		Code:   code,
		Detail: strings.Join(detail, ", "),
	}
}

func ToErrorResponse(e *jsonapi.ErrorObject) *response.ErrorResponse {
	status, err := strconv.Atoi(e.Status)
	if err != nil {
		panic(err)
	}
	return response.NewErrorResponse(status, e)
}

func invalidQueryParams(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, codeInvalidQueryParams, err.Error())
}

func invalidPayload(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, codeInvalidPayload, err.Error())
}

func forbidden(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusForbidden, codeForbidden, err.Error())
}
