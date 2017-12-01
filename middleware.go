package jargo

import (
	"github.com/google/jsonapi"
	"crushedpixel.net/margo"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/goji/param"
)

// gin middleware injecting the jargo application into the context
func injectApplicationMiddleware(app *Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(application, app)
	}
}

// margo middleware validating JSON API requests
func contentTypeMiddleware(action *Action) margo.HandlerFunc {
	// TODO refine
	return func(c *margo.Context) margo.Response {
		var headerKey string
		if action.method == http.MethodGet {
			headerKey = "Accept"
		} else {
			headerKey = "Content-Type"
		}

		if c.Request.Header.Get(headerKey) != jsonapi.MediaType {
			return margo.NewEmptyResponse(http.StatusNotAcceptable)
		}
		return nil
	}
}

// margo middleware unmarshaling jsonapi-specific query parameters
func fetchParamsMiddleware(action *Action) margo.HandlerFunc {
	return func(c *margo.Context) margo.Response {
		if action.method != http.MethodGet {
			return nil
		}

		fp := &FetchParams{}
		param.Parse(c.Request.URL.Query(), fp)

		// TODO: validate query parameters

		// TODO: validate filter contents

		c.Set(fetchParams, fp)
		return nil
	}
}
