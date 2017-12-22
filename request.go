package jargo

import (
	"errors"
	"github.com/go-pg/pg"
	"crushedpixel.net/jargo/api"
)

var ErrIncludeNotSupported = errors.New("include parameter not supported")

// jsonapi query parameters for fetching data
// (see http://jsonapi.org/format/#fetching)
type QueryParams struct {
	Fields api.FieldSet
	Sort   api.SortFields
	Page   api.Pagination
}

// jsonapi query parameters specific to index actions
// (see http://jsonapi.org/format/#fetching)
type IndexQueryParams struct {
	Filter api.Filters
}

func parseCreateRequest(c *Context) (interface{}, error) {
	return c.Resource().ParsePayload(c.Request.Body)
}

func parseUpdateRequest(c *Context) (interface{}, error) {
	instance, err := c.Resource().SelectById(c.Application().DB, c.Params.ByName("id")).GetValue()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, api.ErrNotFound
		}
		return nil, err
	}

	return c.Resource().ParseUpdatePayload(c.Request.Body, instance)
}
