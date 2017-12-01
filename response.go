package jargo

import (
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"crushedpixel.net/margo"
	"encoding/json"
)

type Response interface {
	margo.Response
	Status() int
	Payload() ([]byte, error)
}

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

type DataResponse struct {
	status  int
	payload jsonapi.Payloader
}

func (r *DataResponse) Status() int {
	return r.status
}

func (r *DataResponse) Payload() ([]byte, error) {
	b, err := json.Marshal(r.payload)
	return b, err
}

func (r *DataResponse) Send(c *gin.Context) {
	sendResponse(c, r)
}

func sendResponse(c *gin.Context, r Response) {
	payload, err := r.Payload()
	if err != nil {
		panic(err)
	}

	c.Header("Content-Type", jsonapi.MediaType)
	c.Status(r.Status())
	c.Writer.Write(payload)
}

func NewErrorResponse(status int, errors ...*jsonapi.ErrorObject) *ErrorResponse {
	return &ErrorResponse{
		status,
		&jsonapi.ErrorsPayload{Errors: errors},
	}
}

func NewDataResponse(status int, payload jsonapi.Payloader) *DataResponse {
	return &DataResponse{
		status,
		payload,
	}
}
