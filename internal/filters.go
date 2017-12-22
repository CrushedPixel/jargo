package internal

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"errors"
)

type Filters struct {
	resource *Resource
	filter   map[string]map[string][]string
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

func (f *Filters) ApplyToQuery(q *orm.Query) {
	for column, options := range f.filter {
		whereOr(q, column, "=", options["EQ"])
		whereOr(q, column, "<>", options["NE"])
		whereOr(q, column, "LIKE", options["LIKE"])
		whereOr(q, column, "<", options["LT"])
		whereOr(q, column, "<=", options["LTE"])
		whereOr(q, column, ">", options["GT"])
		whereOr(q, column, ">=", options["GTE"])
	}
}

func whereOr(q *orm.Query, column string, op string, values []string) {
	if values != nil {
		for _, val := range values {
			q.WhereOr(fmt.Sprintf("%s %s ?", column, op), val)
		}
	}
}

func newFilters(resource *Resource, values map[string]map[string][]string) (*Filters, error) {
	filter := make(map[string]map[string][]string)

	for field, operations := range values {
		var column string
		if field == primaryFieldJsonapiName {
			column = primaryFieldColumn
		} else {
			// find the resourceField with matching jsonapi name
			for _, f := range resource.fields {
				if field != f.definition.name {
					continue
				}

				if f.definition.typ != attribute {
					return nil, errors.New("filtering by relations is not supported")
				}
				if !f.definition.sort {
					return nil, errors.New(fmt.Sprintf(`sorting by "%s" is disabled`, field))
				}

				column = f.definition.column
			}
		}

		if column == "" {
			return nil, errors.New(fmt.Sprintf(`unknown filter parameter: "%s"`, field))
		}

		filter[column] = operations
	}

	filters := &Filters{
		resource: resource,
		filter:   filter,
	}
	return filters, nil
}
