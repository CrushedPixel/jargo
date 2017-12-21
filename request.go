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
	instance, err := c.Resource().ParsePayload(c.Request.Body)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func parseUpdateRequest(c *Context) (interface{}, error) {
	instance, err := c.Resource().SelectById(c.Application().DB, c.Params.ByName("id")).GetValue()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ApiErrNotFound
		}
		return nil, err
	}

	err = c.Resource().ParseUpdatePayload(c.Request.Body, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
