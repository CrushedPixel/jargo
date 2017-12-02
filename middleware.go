package jargo

import (
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

// middleware injecting the jargo controller into the context
func injectControllerMiddleware(cont *Controller) HandlerFunc {
	return func(c *Context) Response {
		c.Set(controller, cont)
		return nil
	}
}

// margo middleware validating JSON API requests
func contentTypeMiddleware(action *Action) HandlerFunc {
	return func(c *Context) Response {
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

type parserFetchParams struct {
	Include string            `param:"include"`
	Fields  map[string]string `param:"fields"`
	Filter  map[string]string `param:"filter"`
	Page    map[string]string `param:"page"`
	Sort    string            `param:"sort"`
}

// margo middleware unmarshaling jsonapi-specific query parameters
func fetchParamsMiddleware(action *Action) HandlerFunc {
	return func(c *Context) Response {
		if action.method != http.MethodGet {
			return nil
		}

		pfp := &parserFetchParams{}
		param.Parse(c.Request.URL.Query(), pfp)

		// parse filter settings
		filters, err := parseFilterParameters(c.GetController().Model, pfp.Filter)
		if err != nil {
			return ToErrorResponse(invalidQueryParams(err))
		}

		// parse sort settings
		sorting, err := parseSortParameters(c.GetController().Model, pfp.Sort)
		if err != nil {
			return ToErrorResponse(invalidQueryParams(err))
		}

		fp := &FetchParams{
			Filter: *filters,
			Sort:   *sorting,
		}

		c.Set(fetchParams, fp)
		return nil
	}
}
