package internal

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"errors"
	"net/url"
	"crushedpixel.net/jargo/internal/parser"
)

const (
	dirAscending  = "ASC"
	dirDescending = "DESC"
)

var (
	errSortingByHasRelation       = errors.New("sorting by has relations is not supported")
	errSortingByMany2ManyRelation = errors.New("sorting by many2many relations is not supported")
)

type sortFields struct {
	resource *resource
	sort     map[field]bool // field to order mapping
}

func (s *sortFields) ApplyToQuery(q *orm.Query) {
	for field, asc := range s.sort {
		var dir string
		if asc {
			dir = dirAscending
		} else {
			dir = dirDescending
		}

		q.Order(fmt.Sprintf("%s %s", field.pgFilterColumn(), dir))
	}
}

// parses a comma-separated list of sort fields into a SortFields instance for a given resource
func parseSortFields(resource *resource, query url.Values) (*sortFields, error) {
	fields := parser.ParseSortParameters(query)

	sort := make(map[field]bool)
	for _, fieldName := range fields {
		if len(fieldName) < 1 {
			continue
		}

		// if parameter is prefixed with '-', order is descending
		var asc bool
		if fieldName[0] == '-' {
			asc = false
			fieldName = fieldName[1:]
		} else {
			asc = true
		}

		// find resource field with matching jsonapi name
		var field field
		for _, rf := range resource.fields {
			if rf.jsonapiName() == fieldName {
				if _, ok := rf.(*hasField); ok {
					return nil, errSortingByHasRelation
				}
				// TODO: ensure not sorting by many2many
				field = rf
				break
			}
		}

		if field == nil {
			return nil, errors.New(fmt.Sprintf(`unknown sort parameter: "%s"`, fieldName))
		}
		if !field.sortable() {
			return nil, errors.New(fmt.Sprintf(`sorting by "%s" is disabled`, fieldName))
		}
		sort[field] = asc
	}

	return &sortFields{
		resource: resource,
		sort:     sort,
	}, nil
}
