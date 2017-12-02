package response

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

func sendResponse(c *gin.Context, r Response) {
	payload, err := r.Payload()
	if err != nil {
		panic(err)
	}

	c.Header("Content-Type", jsonapi.MediaType)
	c.Status(r.Status())
	c.Writer.Write(payload)
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

func NewDataResponse(status int, payload jsonapi.Payloader) *DataResponse {
	return &DataResponse{
		status,
		payload,
	}
}
