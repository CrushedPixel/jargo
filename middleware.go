package jargo

import (
	"github.com/gin-gonic/gin"
	"crushedpixel.net/margo"
	"github.com/google/jsonapi"
	"net/http"
	"strings"
	"crushedpixel.net/jargo/api"
)

// gin middleware injecting the jargo application into the context
func injectApplicationMiddleware(app *Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		setApplication(c, app)
	}
}

// middleware injecting the jargo controller into the context
func injectControllerMiddleware(cont *Controller) HandlerFunc {
	return func(c *Context) margo.Response {
		c.setController(cont)
		return nil
	}
}

// margo middleware validating JSON API requests
func contentTypeMiddleware(c *Context) margo.Response {
	// if Content-Type header not the jsonapi media type,
	// return 415 Unsupported Media Type
	ct := c.Request.Header.Get("Content-Type")
	if ct != jsonapi.MediaType &&
		c.Request.Method != http.MethodGet &&
		c.Request.Method != http.MethodDelete {
		return api.ErrUnsupportedMediaType
	}

	var contains, exact bool
	for _, accept := range c.Request.Header["Accept"] {
		if jsonapi.MediaType == accept {
			exact = true
			break
		}

		if strings.Contains(accept, jsonapi.MediaType) {
			contains = true
		}
	}

	// if accept header contains media type but never unmodified,
	// return 406 Not Acceptable
	if contains && !exact {
		return api.ErrNotAcceptable
	}

	return nil
}
