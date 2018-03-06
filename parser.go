package jargo

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	fieldsParamRegex = regexp.MustCompile(`^fields\[([^][]+)]$`)
	filterParamRegex = regexp.MustCompile(`^filter\[([^][]+)](?:\[([^][]+)])?$`)
	pageParamRegex   = regexp.MustCompile(`^page\[([^][]+)]$`)
)

func ParseFieldParameters(query map[string][]string) map[string][]string {
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

func ParseFilterParameters(query map[string][]string) map[string]map[string][]string {
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

func ParsePageParameters(query map[string][]string) map[string]string {
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

func ParseSortParameters(query map[string][]string) map[string]bool {
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

func ParseResourceId(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		return 0, ErrInvalidId
	}
	return id, nil
}

func ParseIndexRequest(base *Request) (*IndexRequest, error) {
	fieldSet, err := base.Resource().ParseFieldSet(ParseFieldParameters(base.QueryParams()))
	if err != nil {
		return nil, err
	}

	filters, err := base.Resource().ParseFilters(ParseFilterParameters(base.QueryParams()))
	if err != nil {
		return nil, err
	}

	pagination, err := base.Resource().ParsePagination(base.Application(),
		ParseSortParameters(base.QueryParams()), ParsePageParameters(base.QueryParams()))

	if err != nil {
		return nil, err
	}

	req := &IndexRequest{
		Request:    base,
		fields:     fieldSet,
		filters:    filters,
		pagination: pagination,
	}
	return req, nil
}

func ParseShowRequest(base *Request) (*ShowRequest, error) {
	fieldSet, err := base.Resource().ParseFieldSet(ParseFieldParameters(base.QueryParams()))
	if err != nil {
		return nil, err
	}

	resourceId, err := ParseResourceId(base.PathParams()["id"])
	if err != nil {
		return nil, err
	}

	req := &ShowRequest{
		Request:    base,
		fields:     fieldSet,
		resourceId: resourceId,
	}
	return req, nil
}

func ParseCreateRequest(base *Request) (*CreateRequest, error) {
	fieldSet, err := base.Resource().ParseFieldSet(ParseFieldParameters(base.QueryParams()))
	if err != nil {
		return nil, err
	}

	req := &CreateRequest{
		Request: base,
		fields:  fieldSet,
	}
	return req, nil
}

func ParseUpdateRequest(base *Request) (*UpdateRequest, error) {
	fieldSet, err := base.Resource().ParseFieldSet(ParseFieldParameters(base.QueryParams()))
	if err != nil {
		return nil, err
	}

	resourceId, err := ParseResourceId(base.PathParams()["id"])
	if err != nil {
		return nil, err
	}

	req := &UpdateRequest{
		Request:    base,
		fields:     fieldSet,
		resourceId: resourceId,
	}
	return req, nil
}

func ParseDeleteRequest(base *Request) (*DeleteRequest, error) {
	resourceId, err := ParseResourceId(base.PathParams()["id"])
	if err != nil {
		return nil, err
	}

	req := &DeleteRequest{
		Request:    base,
		resourceId: resourceId,
	}
	return req, nil
}
