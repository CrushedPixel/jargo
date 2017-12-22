package internal

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"errors"
)

const (
	dirAscending  = "ASC"
	dirDescending = "DESC"
)

type SortFields struct {
	resource *Resource
	sort     map[string]bool // column to order mapping
}

func (s *SortFields) ApplyToQuery(q *orm.Query) {
	for column, asc := range s.sort {
		var dir string
		if asc {
			dir = dirAscending
		} else {
			dir = dirDescending
		}

		q.Order(fmt.Sprintf("%s.%s %s", s.resource.definition.alias, column, dir))
	}
}

// parses a comma-separated list of sort fields into a SortFields instance for a given resource
func newSortFields(resource *Resource, fields []string) (*SortFields, error) {
	sort := make(map[string]bool)

	for _, field := range fields {
		if len(field) < 1 {
			continue
		}

		// if parameter is prefixed with '-', order is descending
		var asc bool
		if field[0] == '-' {
			asc = false
			field = field[1:]
		} else {
			asc = true
		}

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
					return nil, errors.New("sorting by relations is not supported")
				}
				if !f.definition.sort {
					return nil, errors.New(fmt.Sprintf(`sorting by "%s" is disabled`, field))
				}

				column = f.definition.column
			}
		}

		if column == "" {
			return nil, errors.New(fmt.Sprintf(`unknown sort parameter: "%s"`, field))
		}

		sort[column] = asc
	}

	sortFields := &SortFields{
		resource: resource,
		sort:     sort,
	}
	return sortFields, nil
}
