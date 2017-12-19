package application

import (
	"github.com/google/jsonapi"
	"strconv"
	"github.com/satori/go.uuid"
	"github.com/gin-gonic/gin"
)

type ApiError struct {
	Status int
	Code   string
	Detail string
}

// implements error
func (err *ApiError) Error() string {
	return err.Detail
}

// implements margo.Response
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
