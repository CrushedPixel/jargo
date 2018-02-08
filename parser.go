package jargo

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	fieldsParamRegex = regexp.MustCompile(`^fields\[([^][]+)]$`)
	filterParamRegex = regexp.MustCompile(`^filter\[([^][]+)](?:\[([^][]+)])?$`)
	pageParamRegex   = regexp.MustCompile(`^page\[([^][]+)]$`)
)

func parseFieldParameters(query url.Values) map[string][]string {
	fields := make(map[string][]string)
	for k, v := range query {
		res := fieldsParamRegex.FindAllStringSubmatch(k, -1)
		if res == nil {
			continue
		}
		groups := res[0]
		field := groups[1]

		// parse string-separated values
		values := make([]string, 0)
		for _, val := range v {
			values = append(values, strings.Split(val, ",")...)
		}

		fields[field] = values
	}

	return fields
}

func parseFilterParameters(query url.Values) map[string]map[string][]string {
	// map[field]map[operator][]values
	filters := make(map[string]map[string][]string)
	for k, v := range query {
		res := filterParamRegex.FindAllStringSubmatch(k, -1)
		if res == nil {
			continue
		}

		// there's at most one match per key
		groups := res[0]

		field := groups[1]
		op := groups[2]
		if op == "" {
			op = "EQ"
		}
		op = strings.ToUpper(op)

		// add values to operations
		operations := filters[field]
		if operations == nil {
			operations = make(map[string][]string)
		}

		values := make([]string, 0)
		for _, val := range v {
			values = append(values, strings.Split(val, ",")...)
		}
		operations[op] = values

		filters[field] = operations
	}

	return filters
}

func parsePageParameters(query url.Values) map[string]string {
	fields := make(map[string]string)
	for k, v := range query {
		res := pageParamRegex.FindAllStringSubmatch(k, -1)
		if res == nil {
			continue
		}
		groups := res[0]
		fields[groups[1]] = v[0]
	}

	return fields
}

func parseSortParameters(query url.Values) map[string]bool {
	values := make(map[string]bool)
	if sort, ok := query["sort"]; ok {
		for _, v := range sort {
			for _, fieldName := range strings.Split(v, ",") {
				// skip empty sort parameters
				if len(fieldName) < 1 {
					continue
				}

				asc := true
				if fieldName[0] == '-' {
					// if parameter is prefixed with a hyphen,
					// order is descending
					asc = false
					fieldName = fieldName[1:]
				}

				values[fieldName] = asc
			}
		}
	}

	return values
}

func parseResourceId(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		return 0, ErrInvalidId
	}
	return id, nil
}

func parseIndexRequest(c *Context) (*IndexRequest, error) {
	fieldSet, err := c.Resource().ParseFieldSet(parseFieldParameters(c.QueryParams))
	if err != nil {
		return nil, err
	}

	filters, err := c.Resource().ParseFilters(parseFilterParameters(c.QueryParams))
	if err != nil {
		return nil, err
	}

	sort, err := c.Resource().ParseSortFields(parseSortParameters(c.QueryParams))
	if err != nil {
		return nil, err
	}

	pagination, err := ParsePagination(parsePageParameters(c.QueryParams), c.Application().MaxPageSize())
	if err != nil {
		return nil, err
	}

	req := &IndexRequest{
		Fields:     fieldSet,
		Filters:    filters,
		SortFields: sort,
		Pagination: pagination,
	}
	return req, nil
}

func parseShowRequest(c *Context) (*ShowRequest, error) {
	fieldSet, err := c.Resource().ParseFieldSet(parseFieldParameters(c.QueryParams))
	if err != nil {
		return nil, err
	}

	resourceId, err := parseResourceId(c.PathParams["id"])
	if err != nil {
		return nil, err
	}

	req := &ShowRequest{
		Fields:     fieldSet,
		ResourceId: resourceId,
	}
	return req, nil
}

func parseCreateRequest(c *Context) (*CreateRequest, error) {
	fieldSet, err := c.Resource().ParseFieldSet(parseFieldParameters(c.QueryParams))
	if err != nil {
		return nil, err
	}

	req := &CreateRequest{
		Fields:  fieldSet,
		Payload: c.Payload,
	}
	return req, nil
}

func parseUpdateRequest(c *Context) (*UpdateRequest, error) {
	fieldSet, err := c.Resource().ParseFieldSet(parseFieldParameters(c.QueryParams))
	if err != nil {
		return nil, err
	}

	resourceId, err := parseResourceId(c.PathParams["id"])
	if err != nil {
		return nil, err
	}

	req := &UpdateRequest{
		Fields:     fieldSet,
		ResourceId: resourceId,
		Payload:    c.Payload,
	}
	return req, nil
}

func parseDeleteRequest(c *Context) (*DeleteRequest, error) {
	resourceId, err := parseResourceId(c.PathParams["id"])
	if err != nil {
		return nil, err
	}

	req := &DeleteRequest{
		ResourceId: resourceId,
	}
	return req, nil
}
