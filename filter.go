package jargo

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"github.com/go-pg/pg/types"
)

type Filter map[types.Q]*FilterOptions

type FilterOptions struct {
	Eq   []string
	Ne   []string
	Like []string
	Gt   []string
	Gte  []string
	Lt   []string
	Lte  []string
}

func (filter Filter) ApplyToQuery(q *orm.Query) {
	for field, options := range filter {
		whereOr(q, field, "=", options.Eq)
		whereOr(q, field, "<>", options.Ne)
		whereOr(q, field, "LIKE", options.Like)
		whereOr(q, field, "<", options.Lt)
		whereOr(q, field, "<=", options.Lte)
		whereOr(q, field, ">", options.Gt)
		whereOr(q, field, ">=", options.Gte)
	}
}

func whereOr(q *orm.Query, column types.Q, op string, values []string) {
	for _, val := range values {
		q.WhereOr(fmt.Sprintf("%s %s ?", column, op), val)
	}
}
