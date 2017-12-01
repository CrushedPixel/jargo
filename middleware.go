package jargo

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/goji/param"
	"strings"
	"fmt"
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

		// parse and validate filters
		filters := make(Filter)

		// convert filter parameters to FilterOptions struct
		for k, v := range pfp.Filter {
			ks := strings.SplitN(k, ":", 2)

			if len(ks) < 1 {
				continue
			}

			key := ks[0] // the field to filter by

			column, ann := sqlColumnForAttrName(c.GetController().Model, key)
			if column == nil {
				return ToErrorResponse(InvalidFilterError(fmt.Sprintf("filtering by %s is disabled", key)))
			}

			// check if field has jargo:"filter" tag
			if !ann.filter {
				return ToErrorResponse(InvalidFilterError(fmt.Sprintf("filtering by %s is disabled", key)))
			}

			var op string // the filtering operator
			if len(ks) == 2 {
				op = ks[1]
			} else {
				op = "eq"
			}

			values := strings.Split(v, ",")

			filter, ok := filters[*column]
			if !ok {
				filter = &FilterOptions{}
				filters[*column] = filter
			}

			switch op {
			case "eq":
				filter.Eq = append(filter.Eq, values...)
				break
			case "ne":
				filter.Ne = append(filter.Ne, values...)
				break
			case "like":
				filter.Like = append(filter.Like, values...)
				break
			case "gt":
				filter.Gt = append(filter.Gt, values...)
				break
			case "gte":
				filter.Gte = append(filter.Gte, values...)
				break
			case "lt":
				filter.Lt = append(filter.Lt, values...)
				break
			case "lte":
				filter.Lte = append(filter.Lte, values...)
				break
			default:
				return ToErrorResponse(InvalidFilterError(fmt.Sprintf("invalid filter operator: %s", op)))
			}
		}

		fp := &FetchParams{
			Include: pfp.Include,
			Fields:  pfp.Fields,
			Filter:  filters,
			Page:    pfp.Page,
			Sort:    pfp.Sort,
		}

		c.Set(fetchParams, fp)
		return nil
	}
}
