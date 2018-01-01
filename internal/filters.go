package internal

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"errors"
	"net/url"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/jargo/internal/parser"
)

var errFilteringByHasRelation = errors.New("filtering by has relations is not supported")

type filters struct {
	resource *resource
	filter   map[field]map[string][]string
}

type filterOptions struct {
	Eq   []string
	Ne   []string
	Like []string
	Gt   []string
	Gte  []string
	Lt   []string
	Lte  []string
}

func (f *filters) ApplyToQuery(q *orm.Query) {
	for field, options := range f.filter {
		whereOr(q, field, "=", options["EQ"])
		whereOr(q, field, "<>", options["NE"])
		whereOr(q, field, "LIKE", options["LIKE"])
		whereOr(q, field, "<", options["LT"])
		whereOr(q, field, "<=", options["LTE"])
		whereOr(q, field, ">", options["GT"])
		whereOr(q, field, ">=", options["GTE"])
	}
}

func whereOr(q *orm.Query, field field, op string, values []string) {
	if values != nil {
		for _, val := range values {
			q.WhereOr(fmt.Sprintf("%s %s ?", field.pgFilterColumn(), op), val)
		}
	}
}

func parseFilters(resource *resource, query url.Values) (api.Filters, error) {
	parsed, err := parser.ParseFilterParameters(query)
	if err != nil {
		return nil, err
	}

	filter := make(map[field]map[string][]string)
	for fieldName, operations := range parsed {
		// find resource field with matching jsonapi name
		var field field
		for _, rf := range resource.fields {
			if rf.jsonapiName() == fieldName {
				if _, ok := rf.(*hasField); ok {
					return nil, errFilteringByHasRelation
				}
				// TODO: ensure not filtering by many2many
				field = rf
				break
			}
		}

		if field == nil {
			return nil, errors.New(fmt.Sprintf(`unknown filter parameter: "%s"`, fieldName))
		}
		if !field.filterable() {
			return nil, errors.New(fmt.Sprintf(`filtering by "%s" is disabled`, fieldName))
		}
		filter[field] = operations
	}

	filters := &filters{
		resource: resource,
		filter:   filter,
	}
	return filters, nil
}
