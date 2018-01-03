package internal

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"errors"
	"net/url"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/jargo/internal/parser"
	"strconv"
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
		andWhereOr(q, field, "=", options["EQ"])
		andWhereOr(q, field, "<>", options["NE"])
		andWhereOr(q, field, "LIKE", options["LIKE"])
		andWhereOr(q, field, "<", options["LT"])
		andWhereOr(q, field, "<=", options["LTE"])
		andWhereOr(q, field, ">", options["GT"])
		andWhereOr(q, field, ">=", options["GTE"])
	}
}

// generates an AND WHERE (xxx OR xxx) clause
func andWhereOr(q *orm.Query, field field, op string, values []string) {
	if values != nil {
		q.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, val := range values {
				// go-pg does not escape the fields in where clauses,
				// so we need to do it ourselves
				f := escapePGColumn(field.pgFilterColumn())
				q = q.WhereOr(fmt.Sprintf("%s %s ?", f, op), val)
			}
			return q, nil
		})
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

// returns an api.Filters instance filtering by id
func idFilter(resource *resource, id int64) api.Filters {
	var idFld field
	for _, field := range resource.fields {
		if _, ok := field.(*idField); ok {
			idFld = field
		}
	}
	if idFld == nil {
		panic("id field not found")
	}

	filter := map[field]map[string][]string{idFld: {"EQ": {strconv.FormatInt(id, 10)}}}
	return &filters{
		resource: resource,
		filter:   filter,
	}
}
