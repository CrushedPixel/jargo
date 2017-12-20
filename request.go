package jargo

import (
	"github.com/goji/param"
	"errors"
	"github.com/go-pg/pg"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/jargo/internal/parser"
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

func parseQueryParams(c *Context) (*QueryParams, error) {
	if _, ok := c.GetQuery("include"); ok {
		return nil, ErrIncludeNotSupported
	}

	// parse fields parameter
	fields, err := parser.ParseFieldParameters(c.Request.URL.Query())
	if err != nil {
		return nil, err
	}

	// parse sort settings
	sorting, err := parseSortParameters(c.GetController().Resource, parsed.Sort)
	if err != nil {
		return nil, err
	}

	// parse pagination settings
	pagination, err := parsePageParameters(c.GetApplication(), parsed.Page)
	if err != nil {
		return nil, err
	}

	params := &QueryParams{
		Sort:   *sorting,
		Page:   *pagination,
		Fields: fields,
	}

	return params, nil
}

func parseIndexQueryParams(c *Context) (*IndexQueryParams, error) {
	// TODO: change filter query parameter format
	parsed := &parserIndexQueryParams{}
	param.Parse(c.Request.URL.Query(), parsed)

	// parse filter settings
	filters, err := parseFilterParameters(c.GetController().Resource, parsed.Filter)
	if err != nil {
		return nil, err
	}

	params := &IndexQueryParams{
		Filter: *filters,
	}

	return params, nil
}

func parseCreateRequest(c *Context) (interface{}, error) {
	resource := c.GetController().Resource
	instance, err := resource.UnmarshalCreate(c.Request.Body)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func parseUpdateRequest(c *Context) (interface{}, error) {
	res := c.GetController().Resource

	id := c.Params.ByName("id")

	q := res.SelectOne(c.GetApplication().DB)
	q.Raw().Where("id = ?", id)
	val, err := q.GetValue()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ApiErrNotFound
		}
		return nil, err
	}

	instance, err := res.UnmarshalUpdate(c.Request.Body, val, id)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
