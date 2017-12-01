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
)

func NewErrorObject(status int, code string, detail ...string) *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(status),
		Code:   code,
		Detail: strings.Join(detail, ", "),
	}
}

func NewInvalidQueryParamsError() *jsonapi.ErrorObject {
	return NewErrorObject(http.StatusBadRequest, invalidQueryParams)
}