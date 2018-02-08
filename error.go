package jargo

import (
	"bytes"
	"fmt"
	"github.com/google/jsonapi"
	"github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

// An ApiError is a struct containing
// information about an error that occurred
// handling a request.
// ApiError implements error and Response,
// so it can be returned both as error value in
// functions and as Response value in HandlerFuncs.
type ApiError struct {
	status int
	code   string
	detail string
}

// Error satisfies the error interface.
func (e *ApiError) Error() string {
	return e.detail
}

// Status satisfies the Response interface.
func (e *ApiError) Status() int {
	return e.status
}

// Payload satisfies the Response interface.
func (e *ApiError) Payload() (string, error) {
	buf := new(bytes.Buffer)
	err := jsonapi.MarshalErrors(buf, []*jsonapi.ErrorObject{e.ToErrorObject()})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Body satisfies the ferry.Response interface.
func (e *ApiError) Response() (int, string) {
	return responseToFerry(e).Response()
}

// ToErrorObject converts the ApiError to a jsonapi.ErrorObject.
func (e *ApiError) ToErrorObject() *jsonapi.ErrorObject {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return &jsonapi.ErrorObject{
		ID:     u.String(),
		Status: strconv.Itoa(e.status),
		Code:   e.code,
		Detail: e.detail,
	}
}

// NewApiError returns a new ApiError from a status code,
// error code and error detail string.
func NewApiError(status int, code string, detail string) *ApiError {
	return &ApiError{
		status: status,
		code:   code,
		detail: detail,
	}
}

// NewErrorResponse returns a Response containing
// an error payload according to the JSON API spec.
// See http://jsonapi.org/format/#errors
//
// If the underlying error is an instance of ApiError,
// the error itself is returned.
// Otherwise, it logs the error as an internal server error
// and returns ErrInternalServerError.
func NewErrorResponse(err error) *ApiError {
	apiError, ok := err.(*ApiError)
	if !ok {
		// if error is not an api error, return internal server error response
		println(fmt.Sprintf("Internal server error: %s", err.Error())) // TODO use a proper logging library
		apiError = ErrInternalServerError
	}

	return apiError
}

// ErrInternalServerError is an ApiError indicating
// an unspecified internal error.
var ErrInternalServerError = NewApiError(
	http.StatusInternalServerError,
	"INTERNAL_SERVER_ERROR",
	"internal server error",
)

var ErrUnsupportedMediaType = NewApiError(
	http.StatusUnsupportedMediaType,
	"UNSUPPORTED_MEDIA_TYPE",
	fmt.Sprintf("media type must be %s", jsonapi.MediaType),
)

var ErrNotAcceptable = NewApiError(
	http.StatusNotAcceptable,
	"NOT_ACCEPTABLE",
	fmt.Sprintf("accept header must contain %s without any media type parameters", jsonapi.MediaType),
)

var ErrNotFound = NewApiError(
	http.StatusNotFound,
	"RESOURCE_NOT_FOUND",
	"resource not found",
)

var ErrForbidden = NewApiError(
	http.StatusForbidden,
	"FORBIDDEN",
	"forbidden",
)

var ErrInvalidId = NewApiError(
	http.StatusBadRequest,
	"INVALID_ID",
	"invalid id parameter",
)

func ErrUnauthorized(detail string) *ApiError {
	return NewApiError(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		detail,
	)
}

func ErrInvalidQueryParams(detail string) *ApiError {
	return NewApiError(http.StatusBadRequest,
		"INVALID_QUERY_PARAMS",
		detail,
	)
}

func ErrInvalidPayload(detail string) *ApiError {
	return NewApiError(http.StatusBadRequest,
		"INVALID_PAYLOAD",
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
