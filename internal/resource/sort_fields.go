package resource

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"strings"
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

		q.Order(fmt.Sprintf("%s %s", column, dir))
	}
}

// parses a comma-separated list of sort fields into a SortFields instance for a given resource
func ParseSortFields(resource *Resource, val string) (*SortFields, error) {
	sort := make(map[string]bool)

	if val != "" {
		spl := strings.Split(val, ",")
		for _, key := range spl {
			if len(key) < 1 {
				return nil, errors.New(fmt.Sprintf(`invalid sort parameter: "%s"`, key))
			}

			// if parameter is prefixed with '-', order is descending
			var asc bool
			if key[0] == '-' {
				asc = false
				key = key[1:]
			} else {
				asc = true
			}

			var column string
			if key == primaryFieldJsonapiName {
				column = primaryFieldColumn
			} else {
				// find the resourceField with matching jsonapi name
				for _, f := range resource.fields {
					if f.definition.typ != attribute {
						return nil, errors.New("sorting by relations is not supported")
					}
					if !f.definition.sort {
						return nil, errors.New(fmt.Sprintf(`sorting by "%s" is disabled`, key))
					}

					if key == f.definition.name {
						column = f.definition.column
					}
				}
			}

			if column == "" {
				return nil, errors.New(fmt.Sprintf(`unknown sort parameter: "%s"`, key))
			}

			sort[column] = asc
		}
	}

	sortFields := &SortFields{
		resource: resource,
		sort:     sort,
	}
	return sortFields, nil
}
