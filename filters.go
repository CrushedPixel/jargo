package jargo

import (
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
)

type Filters struct {
	resource *Resource
	filter   map[internal.SchemaField]map[string][]string
}

func newFilters(r *Resource, filter map[internal.SchemaField]map[string][]string) *Filters {
	return &Filters{
		resource: r,
		filter:   filter,
	}
}

func (f *Filters) applyToQuery(q *orm.Query) {
	for field, options := range f.filter {
		andWhereOr(q, field, "=", options["EQ"])
		andWhereOr(q, field, "<>", options["NOT"])
		andWhereOr(q, field, "LIKE", options["LIKE"])
		andWhereOr(q, field, "<", options["LT"])
		andWhereOr(q, field, "<=", options["LTE"])
		andWhereOr(q, field, ">", options["GT"])
		andWhereOr(q, field, ">=", options["GTE"])
	}
}

// generates an AND WHERE (xxx OR xxx) clause
func andWhereOr(q *orm.Query, field internal.SchemaField, op string, values []string) {
	if values != nil {
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
