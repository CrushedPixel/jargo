package response

import (
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"encoding/json"
)

type ErrorResponse struct {
	status  int
	payload *jsonapi.ErrorsPayload
}

func (r *ErrorResponse) Status() int {
	return r.status
}

func (r *ErrorResponse) Payload() ([]byte, error) {
	b, err := json.Marshal(r.payload)
	return b, err
}

func (r *ErrorResponse) Send(c *gin.Context) {
	sendResponse(c, r)
}

func NewErrorResponse(status int, errors ...*jsonapi.ErrorObject) *ErrorResponse {
	return &ErrorResponse{
		status,
		&jsonapi.ErrorsPayload{Errors: errors},
	}
}