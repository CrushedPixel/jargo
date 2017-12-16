package jargo

import (
	"net/http"
	"fmt"
	"github.com/google/jsonapi"
)

const (
	codeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	codeNotAcceptable        = "NOT_ACCEPTABLE"
	codeInternalServerError  = "INTERNAL_SERVER_ERROR"
	codeNotFound             = "RESOURCE_NOT_FOUND"
	codeInvalidQueryParams   = "INVALID_QUERY_PARAMS"
	codeInvalidPayload       = "INVALID_PAYLOAD"
	codeUnauthorized         = "UNAUTHORIZED"
	codeForbidden            = "FORBIDDEN"
)

var ApiErrInternalServerError = NewApiError(
	http.StatusInternalServerError,
	codeInternalServerError,
	"internal server error",
)

var ApiErrUnsupportedMediaType = NewApiError(
	http.StatusUnsupportedMediaType,
	codeUnsupportedMediaType,
	fmt.Sprintf("media type must be %s", jsonapi.MediaType),
)

var ApiErrNotAcceptable = NewApiError(
	http.StatusNotAcceptable,
	codeNotAcceptable,
	fmt.Sprintf("accept header must contain %s without any media type parameters", jsonapi.MediaType),
)

var ApiErrNotFound = NewApiError(
	http.StatusNotFound,
	codeNotFound,
	"resource not found",
)

var ApiErrForbidden = NewApiError(
	http.StatusForbidden,
	codeForbidden,
	"forbidden",
)

func ApiErrUnauthorized(detail string) *ApiError {
	return NewApiError(
		http.StatusUnauthorized,
		codeUnauthorized,
		detail,
	)
}

func ApiErrInvalidQueryParams(detail string) *ApiError {
	return NewApiError(http.StatusBadRequest,
		codeInvalidQueryParams,
		detail,
	)
}

func ApiErrInvalidPayload(detail string) *ApiError {
	return NewApiError(http.StatusBadRequest,
		codeInvalidPayload,
		detail,
	)
}
