package jargo

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/goji/param"
)

// gin middleware injecting the jargo application into the context
func injectApplicationMiddleware(app *Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		setApplication(c, app)
	}
}

// middleware injecting the jargo controller into the context
func injectControllerMiddleware(cont *Controller) HandlerFunc {
	return func(c *Context) Response {
		c.setController(cont)
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

// middleware unmarshaling jsonapi-specific query parameters
func parseRequestMiddleware(action *Action) HandlerFunc {
	switch action.method {
	case http.MethodGet:
		return parseFetchRequest
	case http.MethodPost:
		return parseCreateRequest
	default:
		return nil
	}
}

func parseFetchRequest(c *Context) Response {
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

	// parse pagination settings
	pagination, err := parsePageParameters(c.GetApplication(), pfp.Page)
	if err != nil {
		return ToErrorResponse(invalidQueryParams(err))
	}

	fp := &FetchParams{
		Filter: *filters,
		Sort:   *sorting,
		Page:   *pagination,
	}

	c.setFetchParams(fp)
	return nil
}

func parseCreateRequest(c *Context) Response {
	model := c.GetController().Model
	instance, err := model.UnmarshalCreate(c.Request.Body)
	if err != nil {
		return ToErrorResponse(invalidPayload(err))
	}

	c.setCreatedModel(instance)
	return nil
}
