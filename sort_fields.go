package jargo

import (
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
)

const (
	dirAscending  = "ASC"
	dirDescending = "DESC"
)

type SortFields struct {
	resource *Resource
	sort     map[internal.SchemaField]bool // field to order mapping
}

func newSortFields(r *Resource, sort map[internal.SchemaField]bool) *SortFields {
	return &SortFields{
		resource: r,
		sort:     sort,
	}
}

func (s *SortFields) applyToQuery(q *orm.Query) {
	for field, asc := range s.sort {
		var dir string
		if asc {
			dir = dirAscending
		} else {
			dir = dirDescending
		}

		q.Order(fmt.Sprintf("%s %s", field.PGFilterColumn(), dir))
	}
}
