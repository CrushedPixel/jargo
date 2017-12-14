package jargo

import (
	"github.com/goji/param"
	"errors"
)

var ErrIncludeNotSupported = errors.New("include parameter not supported")

// JSON API query parameters for fetching data
// (see http://jsonapi.org/format/#fetching)
type QueryParams struct {
	// Include params.Include
	Fields ResultFields
	Sort   ResultSorting
	Page   ResultPagination
}

// used internally to parse jsonapi query parameters
type parserQueryParams struct {
	Include string            `param:"include"`
	Fields  map[string]string `param:"fields"`
	Page    map[string]string `param:"page"`
	Sort    string            `param:"sort"`
}

// jsonapi query parameters specific to index actions
// (see http://jsonapi.org/format/#fetching)
type IndexQueryParams struct {
	Filter Filters
}

// used internally to parse index action specific query parameters
type parserIndexQueryParams struct {
	Filter map[string]string `param:"filter"`
}

func parseQueryParams(c *Context) (*QueryParams, error) {
	parsed := &parserQueryParams{}
	param.Parse(c.Request.URL.Query(), parsed)

	// parse include settings
	if parsed.Include != "" {
		return nil, ErrIncludeNotSupported
	}

	// parse fields settings
	fields, err := parseFieldParameters(parsed.Fields)

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
	/*
	model := c.GetController().Resource

	// TODO find current instance by id

	instance, err := model.UnmarshalUpdate(c.Body)
	if err != nil {
		return nil, err
	}

	return instance, nil
	*/
	return nil, nil
}
