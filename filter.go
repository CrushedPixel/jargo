package jargo

import (
	"github.com/go-pg/pg/orm"
	"fmt"
	"github.com/go-pg/pg/types"
	"strings"
	"errors"
)

type Filters map[types.Q]*FilterOptions

type FilterOptions struct {
	Eq   []string
	Ne   []string
	Like []string
	Gt   []string
	Gte  []string
	Lt   []string
	Lte  []string
}

func (filter *Filters) ApplyToQuery(q *orm.Query) {
	for column, options := range *filter {
		whereOr(q, column, "=", options.Eq)
		whereOr(q, column, "<>", options.Ne)
		whereOr(q, column, "LIKE", options.Like)
		whereOr(q, column, "<", options.Lt)
		whereOr(q, column, "<=", options.Lte)
		whereOr(q, column, ">", options.Gt)
		whereOr(q, column, ">=", options.Gte)
	}
}

func whereOr(q *orm.Query, column types.Q, op string, values []string) {
	for _, val := range values {
		q.WhereOr(fmt.Sprintf("%s %s ?", column, op), val)
	}
}

func parseFilterParameters(model *Model, values map[string]string) (*Filters, error) {
	filters := make(Filters)

	for k, v := range values {
		spl := strings.SplitN(k, ":", 2)

		if len(spl) < 1 {
			continue
		}

		key := spl[0] // the field to filter by

		field, ok := model.Fields[key]
		if !ok {
			return nil, errors.New(fmt.Sprintf("unknown filter attribute: %s", key))
		}

		// check if field has jargo:"filter" tag
		if !field.Settings.AllowFiltering {
			return nil, errors.New(fmt.Sprintf("filtering by %s is disabled", key))
		}

		var op string // the filtering operator
		if len(spl) == 2 {
			op = spl[1]
		} else {
			op = "eq"
		}

		values := strings.Split(v, ",")

		filter, ok := filters[field.Column]
		if !ok {
			filter = &FilterOptions{}
			filters[field.Column] = filter
		}

		switch op {
		case "eq":
			filter.Eq = append(filter.Eq, values...)
			break
		case "ne":
			filter.Ne = append(filter.Ne, values...)
			break
		case "like":
			filter.Like = append(filter.Like, values...)
			break
		case "gt":
			filter.Gt = append(filter.Gt, values...)
			break
		case "gte":
			filter.Gte = append(filter.Gte, values...)
			break
		case "lt":
			filter.Lt = append(filter.Lt, values...)
			break
		case "lte":
			filter.Lte = append(filter.Lte, values...)
			break
		default:
			return nil, errors.New(fmt.Sprintf("invalid filter operator: %s", op))
		}
	}

	return &filters, nil
}
