package response

import (
	"crushedpixel.net/jargo/models"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"github.com/google/jsonapi"
	"net/http"
)

type QueryResponse struct {
	Query *models.Query
}

func NewQueryResponse(q *models.Query) *QueryResponse {
	return &QueryResponse{
		Query: q,
	}
}

func (r *QueryResponse) Status() int {
	if r.Query.Type == models.Insert {
		return http.StatusCreated
	} else {
		return http.StatusOK
	}
}

func (r *QueryResponse) Payload() ([]byte, error) {
	val, err := r.Query.GetValue()
	if err != nil {
		return nil, err
	}

	p, err := jsonapi.Marshal(val)
	if err != nil {
		return nil, err
	}

	return json.Marshal(p)
}

func (r *QueryResponse) Send(c *gin.Context) {
	sendResponse(c, r)
}