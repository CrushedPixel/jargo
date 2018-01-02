package api

import (
	"github.com/google/jsonapi"
	"strconv"
	"github.com/satori/go.uuid"
	"github.com/gin-gonic/gin"
	"fmt"
	"crushedpixel.net/margo"
)

// implements error and margo.Response
type ApiError struct {
	Status int
	Code   string
	Detail string
}

// satisfies error
func (err *ApiError) Error() string {
	return err.Detail
}

// satisfies margo.Response
func (err *ApiError) Send(c *gin.Context) error {
	c.Status(err.Status)
	c.Header("Content-Type", jsonapi.MediaType)
	return jsonapi.MarshalErrors(c.Writer, []*jsonapi.ErrorObject{err.ToErrorObject()})
}

func (err *ApiError) ToErrorObject() *jsonapi.ErrorObject {
	return &jsonapi.ErrorObject{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(err.Status),
		Code:   err.Code,
		Detail: err.Detail,
	}
}

func NewApiError(status int, code string, detail string) *ApiError {
	return &ApiError{
		Status: status,
		Code:   code,
		Detail: detail,
	}
}

// implements margo.Response
type ErrorResponse struct {
	Error error
}

// satisfies margo.Response
func (r *ErrorResponse) Send(c *gin.Context) error {
	apiError, ok := r.Error.(*ApiError)

	if !ok {
		// if error is not an api error, return internal server error response
		println(fmt.Sprintf("Internal server error: %s", r.Error.Error())) // TODO use a proper logging library
		apiError = ErrInternalServerError
	}

	return apiError.Send(c)
}

func NewErrorResponse(err error) margo.Response {
	return &ErrorResponse{Error: err}
}