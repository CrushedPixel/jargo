package jargo

import (
	"fmt"
	"strings"
	"errors"
	"crushedpixel.net/jargo/models"
)

// Custom filter format: filter[name:like]=*name*
// TODO: implement better custom filter format
type Filters map[*models.ModelField]*FilterOptions

type FilterOptions struct {
	Eq   []string
	Ne   []string
	Like []string
	Gt   []string
	Gte  []string
	Lt   []string
	Lte  []string
}

func (filter *Filters) ApplyToQuery(q *models.Query) {
	for field, options := range *filter {
		whereOr(q, field, "=", options.Eq)
		whereOr(q, field, "<>", options.Ne)
		whereOr(q, field, "LIKE", options.Like)
		whereOr(q, field, "<", options.Lt)
		whereOr(q, field, "<=", options.Lte)
		whereOr(q, field, ">", options.Gt)
		whereOr(q, field, ">=", options.Gte)
	}
}

func whereOr(q *models.Query, field *models.ModelField, op string, values []string) {
	for _, val := range values {
		if field.Type == models.RelationField {
			// TODO support filtering by relationship
			println("FILTERING BY RELATIONSHIPS NOT SUPPORTED YET")
		} else {
			q.WhereOr(fmt.Sprintf("%s %s ?", field.PGField.Column, op), val)
		}
	}
}

func parseFilterParameters(model *models.Model, values map[string]string) (*Filters, error) {
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

		filter, ok := filters[field]
		if !ok {
			filter = &FilterOptions{}
			filters[field] = filter
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
