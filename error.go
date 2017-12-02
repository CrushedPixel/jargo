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

func invalidQueryParams(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, codeInvalidQueryParams, err.Error())
}

func invalidPayload(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, codeInvalidPayload, err.Error())
}

func forbidden(err error) *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusForbidden, codeForbidden, err.Error())
}
