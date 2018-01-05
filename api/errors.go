package api

import (
	"net/http"
	"fmt"
	"github.com/google/jsonapi"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

const (
	codeInternalServerError  = "INTERNAL_SERVER_ERROR"
	codeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	codeNotAcceptable        = "NOT_ACCEPTABLE"
	codeNotFound             = "RESOURCE_NOT_FOUND"
	codeInvalidQueryParams   = "INVALID_QUERY_PARAMS"
	codeInvalidPayload       = "INVALID_PAYLOAD"
	codeInvalidId            = "INVALID_ID"
	codeUnauthorized         = "UNAUTHORIZED"
	codeForbidden            = "FORBIDDEN"
)

var ErrInternalServerError = NewApiError(
	http.StatusInternalServerError,
	codeInternalServerError,
	"internal server error",
)

var ErrUnsupportedMediaType = NewApiError(
	http.StatusUnsupportedMediaType,
	codeUnsupportedMediaType,
	fmt.Sprintf("media type must be %s", jsonapi.MediaType),
)

var ErrNotAcceptable = NewApiError(
	http.StatusNotAcceptable,
	codeNotAcceptable,
	fmt.Sprintf("accept header must contain %s without any media type parameters", jsonapi.MediaType),
)

var ErrNotFound = NewApiError(
	http.StatusNotFound,
	codeNotFound,
	"resource not found",
)

var ErrForbidden = NewApiError(
	http.StatusForbidden,
	codeForbidden,
	"forbidden",
)

var ErrInvalidId = NewApiError(
	http.StatusBadRequest,
	codeInvalidId,
	"invalid id parameter",
)

func ErrUnauthorized(detail string) *ApiError {
	return NewApiError(
		http.StatusUnauthorized,
		codeUnauthorized,
		detail,
	)
}

func ErrInvalidQueryParams(detail string) *ApiError {
	return NewApiError(http.StatusBadRequest,
		codeInvalidQueryParams,
		detail,
	)
}

func ErrInvalidPayload(detail string) *ApiError {
	return NewApiError(http.StatusBadRequest,
		codeInvalidPayload,
		detail,
	)
}

func ErrValidationFailed(errors validator.ValidationErrors) *ApiError {
	var failed []string
	for _, v := range errors {
		failed = append(failed, v.Tag())
	}

	// TODO: more descriptive error detail
	return ErrInvalidPayload(strings.Join(failed, ", "))
}
