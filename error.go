package jargo

import (
	"fmt"
	"github.com/crushedpixel/margo"
	"github.com/gin-gonic/gin"
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
// ApiError implements error and margo.Response,
// so it can be returned both as error value in
// functions and as Response value in HandlerFuncs.
type ApiError struct {
	Status int
	Code   string
	Detail string
}

// Error satisfies the error interface.
func (err *ApiError) Error() string {
	return err.Detail
}

// Send writes an error payload to the response body
// according to the JSON API spec.
// See http://jsonapi.org/format/#errors
//
// Satisfies the margo.Response interface.
func (err *ApiError) Send(c *gin.Context) error {
	c.Status(err.Status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalErrors(c.Writer, []*jsonapi.ErrorObject{err.ToErrorObject()})
}

// ToErrorObject converts the ApiError to a jsonapi.ErrorObject.
func (err *ApiError) ToErrorObject() *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(err.Status),
		Code:   err.Code,
		Detail: err.Detail,
	}
}

// NewApiError returns a new ApiError from a status code,
// error code and error detail string.
func NewApiError(status int, code string, detail string) *ApiError {
	return &ApiError{
		Status: status,
		Code:   code,
		Detail: detail,
	}
}

type errorResponse struct {
	Error error
}

func (r *errorResponse) Send(c *gin.Context) error {
	apiError, ok := r.Error.(*ApiError)
	if !ok {
		// if error is not an api error, return internal server error response
		println(fmt.Sprintf("Internal server error: %s", r.Error.Error())) // TODO use a proper logging library
		apiError = ErrInternalServerError
	}

	return apiError.Send(c)
}

// NewErrorResponse returns a margo.Response writing
// an error payload to the response body
// according to the JSON API spec.
// See http://jsonapi.org/format/#errors
//
// If the underlying error is an instance of ApiError,
// sending the Response invokes the ApiError's Send method.
// Otherwise, it logs the ApiError as an internal server error
// and sends ErrInternalServerError.
//
// Satisfies the margo.Response interface.
func NewErrorResponse(err error) margo.Response {
	return &errorResponse{Error: err}
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
