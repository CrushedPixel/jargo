package jargo

import (
	"github.com/goji/param"
	"crushedpixel.net/jargo/models"
)

// JSON API query parameters for fetching data
// (see http://jsonapi.org/format/#fetching)
type FetchParams struct {
	// Include params.Include
	// Fields  params.Fields
	Filter Filters
	Sort   Sorting
	Page   Pagination
}

func (f *FetchParams) ApplyToQuery(q *models.Query) {
	f.Filter.ApplyToQuery(q)
	f.Sort.ApplyToQuery(q)
	f.Page.ApplyToQuery(q)
}

type parserFetchParams struct {
	Include string            `param:"include"`
	Fields  map[string]string `param:"fields"`
	Filter  map[string]string `param:"filter"`
	Page    map[string]string `param:"page"`
	Sort    string            `param:"sort"`
}

func parseFetchRequest(c *Context) (*FetchParams, error) {
	pfp := &parserFetchParams{}
	param.Parse(c.Request.URL.Query(), pfp)

	// parse filter settings
	filters, err := parseFilterParameters(c.GetController().Model, pfp.Filter)
	if err != nil {
		return nil, err
	}

	// parse sort settings
	sorting, err := parseSortParameters(c.GetController().Model, pfp.Sort)
	if err != nil {
		return nil, err
	}

	// parse pagination settings
	pagination, err := parsePageParameters(c.GetApplication(), pfp.Page)
	if err != nil {
		return nil, err
	}

	fp := &FetchParams{
		Filter: *filters,
		Sort:   *sorting,
		Page:   *pagination,
	}

	return fp, nil
}

func parseCreateRequest(c *Context) (interface{}, error) {
	model := c.GetController().Model
	instance, err := model.UnmarshalCreate(c.Request.Body)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func parseUpdateRequest(c *Context) (interface{}, error) {
	/*
	model := c.GetController().Model

	// TODO find current instance by id

	instance, err := model.UnmarshalUpdate(c.Request.Body)
	if err != nil {
		return nil, err
	}

	return instance, nil
	*/
	return nil, nil
}
