package jargo

import (
	"github.com/gin-gonic/gin"
)

// gin middleware injecting the jargo application into the context
func injectApplicationMiddleware(app *Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		setApplication(c, app)
	}
}

// middleware injecting the jargo controller into the context
func injectControllerMiddleware(cont *Controller) HandlerFunc {
	return func(c *Context) interface{} {
		c.setController(cont)
		return nil
	}
}

// margo middleware validating JSON API requests
func contentTypeMiddleware(action *Action) HandlerFunc {
	return func(c *Context) interface{} {
		/* TODO reimplement according to jsonapi spec
		var headerKey string
		if action.method == http.MethodGet {
			headerKey = "Accept"
		} else {
			headerKey = "Content-Type"
		}

		if c.Request.Header.Get(headerKey) != jsonapi.MediaType {
			return margo.NewEmptyResponse(http.StatusNotAcceptable) // TODO
		}
		*/
		return nil
	}
}
