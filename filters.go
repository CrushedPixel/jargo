package jargo

import (
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
)

type Filters struct {
	resource *Resource
	filters  map[internal.SchemaField]*Filter
}

// A Filter contains values to be filtered by,
// each of the filter operators being connected
// via a logical OR, and all of the values for
// an operator being connected via a logical AND.
type Filter struct {
	Eq   []string
	Not  []string
	Like []string
	Lt   []string
	Lte  []string
	Gt   []string
	Gte  []string
}

func newFilters(r *Resource, filters map[internal.SchemaField]*Filter) *Filters {
	return &Filters{
		resource: r,
		filters:  filters,
	}
}

func (f *Filters) applyToQuery(q *orm.Query) {
	for field, filter := range f.filters {
		filter.applyToQuery(q, field)
	}
}

func (f *Filter) applyToQuery(q *orm.Query, field internal.SchemaField) {
	andWhereOr(q, field, "=", f.Eq)
	andWhereOr(q, field, "<>", f.Not)
	andWhereOr(q, field, "LIKE", f.Like)
	andWhereOr(q, field, "<", f.Lt)
	andWhereOr(q, field, "<=", f.Lte)
	andWhereOr(q, field, ">", f.Gt)
	andWhereOr(q, field, ">=", f.Gte)
}

// generates an AND WHERE (xxx OR xxx) clause
func andWhereOr(q *orm.Query, field internal.SchemaField, op string, values []string) {
	if len(values) > 0 {
		q.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, val := range values {
				// go-pg does not escape the fields in where clauses,
				// so we need to do it ourselves
				f := escapePGColumn(field.PGFilterColumn())
				q = q.WhereOr(fmt.Sprintf("%s %s ?", f, op), val)
			}
			return q, nil
		})
	}
}
