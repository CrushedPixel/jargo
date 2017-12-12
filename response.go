package jargo

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"errors"
	"fmt"
	"crushedpixel.net/jargo/models"
	"github.com/google/jsonapi"
	"net/http"
	"strconv"
)

type Response struct {
	value interface{}
}

func NewResponse(value interface{}) *Response {
	return &Response{value: value}
}

// the following value types are turned into the following responses:
// *models.Query        => json-encoded result of query execution
// *jsonapi.ErrorObject => json-encoded error object
// *struct              => json-encoded resource
// []*struct            => json-encoded resource collection
// int                  => HTTP Status Code
// error                => 500 Internal Server Error
// true                 => 204 No Content
// false                => 401 Unauthorized
//
// satisfies margo.Response
func (r *Response) Send(c *gin.Context) {
	value := r.value

	// error is handled by recover() statement in action.go#toMargoHandler
	if err, ok := r.value.(error); ok {
		panic(err)
	}

	// *jsonapi.ErrorObject
	if errObj, ok := r.value.(*jsonapi.ErrorObject); ok {
		i, err := strconv.Atoi(errObj.Status)
		if err != nil {
			panic(err)
		}
		c.Status(i)

		setJsonApiHeaders(c)

		err = jsonapi.MarshalErrors(c.Writer, []*jsonapi.ErrorObject{errObj})
		if err != nil {
			panic(err)
		}
		return
	}

	if query, ok := r.value.(*models.Query); ok {
		var err error
		value, err = query.GetValue()
		if err != nil {
			panic(err)
		}
	}

	val := reflect.ValueOf(value)
	typ := val.Type()

	switch typ.Kind() {
	case reflect.Ptr:
		// *struct
		if typ.Elem().Kind() == reflect.Struct {
			setJsonApiHeaders(c)
			// do not include relationships, as they are not populated with relationships
			err := jsonapi.MarshalPayloadWithoutIncluded(c.Writer, value)
			if err != nil {
				panic(err)
			}
			return
		}

		break
	case reflect.Slice:
		// []*struct
		if typ.Elem().Kind() == reflect.Ptr && typ.Elem().Elem().Kind() == reflect.Struct {
			setJsonApiHeaders(c)
			// do not include relationships, as they are not populated with relationships
			err := jsonapi.MarshalPayloadWithoutIncluded(c.Writer, value)
			if err != nil {
				panic(err)
			}
			return
		}

		break
	case reflect.Int:
		c.Status(int(val.Int()))
		return

	case reflect.Bool:
		if val.Bool() {
			c.Status(http.StatusNoContent)
		} else {
			c.Status(http.StatusForbidden)
		}
		return
	}

	panic(errors.New(fmt.Sprintf("invalid handler return value type: %s", typ)))
}

func setJsonApiHeaders(c *gin.Context) {
	c.Header("Content-Type", jsonapi.MediaType)
}
