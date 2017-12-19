package resource

import (
	"github.com/go-pg/pg/orm"
	"fmt"
)

type Filters struct {
	resource *Resource
	fields   map[*resourceField]*filterOptions
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
	for field, options := range f.fields {
		whereOr(q, field, "=", options.Eq)
		whereOr(q, field, "<>", options.Ne)
		whereOr(q, field, "LIKE", options.Like)
		whereOr(q, field, "<", options.Lt)
		whereOr(q, field, "<=", options.Lte)
		whereOr(q, field, ">", options.Gt)
		whereOr(q, field, ">=", options.Gte)
	}
}

func whereOr(q *orm.Query, field *resourceField, op string, values []string) {
	for _, val := range values {
		q.WhereOr(fmt.Sprintf("%s %s ?", field.definition.column, op), val)
	}
}

// TODO parse filters