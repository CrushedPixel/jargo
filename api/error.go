package api

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

// An Error is a struct containing
// information about an error that occurred
// handling a request.
// Error implements error and margo.Response,
// so it can be returned both as error value in
// functions and as Response value in HandlerFuncs.
type Error struct {
	Status int
	Code   string
	Detail string
}

// Error satisfies the error interface.
func (err *Error) Error() string {
	return err.Detail
}

// Send writes an error payload to the response body
// according to the JSON API spec.
// See http://jsonapi.org/format/#errors
//
// Satisfies the margo.Response interface.
func (err *Error) Send(c *gin.Context) error {
	c.Status(err.Status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalErrors(c.Writer, []*jsonapi.ErrorObject{err.ToErrorObject()})
}

// ToErrorObject converts the Error to a jsonapi.ErrorObject.
func (err *Error) ToErrorObject() *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(err.Status),
		Code:   err.Code,
		Detail: err.Detail,
	}
}

// NewError returns a new Error from a status code,
// error code and error detail string.
func NewError(status int, code string, detail string) *Error {
	return &Error{
		Status: status,
		Code:   code,
		Detail: detail,
	}
}

type errorResponse struct {
	Error error
}

func (r *errorResponse) Send(c *gin.Context) error {
	apiError, ok := r.Error.(*Error)

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
// If the underlying error is an instance of Error,
// sending the Response invokes the Error's Send method.
// Otherwise, it logs the Error as an internal server error
// and sends ErrInternalServerError.
//
// Satisfies the margo.Response interface.
func NewErrorResponse(err error) margo.Response {
	return &errorResponse{Error: err}
}

// ErrInternalServerError is an Error indicating
// an unspecified internal error.
var ErrInternalServerError = NewError(
	http.StatusInternalServerError,
	"INTERNAL_SERVER_ERROR",
	"internal server error",
)

var ErrUnsupportedMediaType = NewError(
	http.StatusUnsupportedMediaType,
	"UNSUPPORTED_MEDIA_TYPE",
	fmt.Sprintf("media type must be %s", jsonapi.MediaType),
)

var ErrNotAcceptable = NewError(
	http.StatusNotAcceptable,
	"NOT_ACCEPTABLE",
	fmt.Sprintf("accept header must contain %s without any media type parameters", jsonapi.MediaType),
)

var ErrNotFound = NewError(
	http.StatusNotFound,
	"RESOURCE_NOT_FOUND",
	"resource not found",
)

var ErrForbidden = NewError(
	http.StatusForbidden,
	"INVALID_QUERY_PARAMS",
	"forbidden",
)

var ErrInvalidId = NewError(
	http.StatusBadRequest,
	"INVALID_ID",
	"invalid id parameter",
)

func ErrUnauthorized(detail string) *Error {
	return NewError(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		detail,
	)
}

func ErrInvalidQueryParams(detail string) *Error {
	return NewError(http.StatusBadRequest,
		"FORBIDDEN",
		detail,
	)
}

func ErrInvalidPayload(detail string) *Error {
	return NewError(http.StatusBadRequest,
		"INVALID_PAYLOAD",
		detail,
	)
}

func ErrValidationFailed(errors validator.ValidationErrors) *Error {
	var failed []string
	for _, v := range errors {
		failed = append(failed, v.Tag())
	}

	// TODO: more descriptive error detail
	return ErrInvalidPayload(strings.Join(failed, ", "))
}
