package jargo

import (
	"fmt"
	"strings"
	"errors"
)

const (
	dirAscending  = "ASC"
	dirDescending = "DESC"
)

type ResultSorting map[*ResourceField]bool

func (sort *ResultSorting) ApplyToQuery(q *Query) {
	for field, asc := range *sort {
		var dir string
		if asc {
			dir = dirAscending
		} else {
			dir = dirDescending
		}

		if field.Type == RelationField {
			// TODO support sorting by relationship
			println("SORTING BY RELATIONSHIPS NOT SUPPORTED YET")
		} else {
			q.OrderExpr(fmt.Sprintf("%s %s", field.PGField.Column, dir))
		}
	}
}

func parseSortParameters(resource *Resource, str string) (*ResultSorting, error) {
	sorting := make(ResultSorting)

	if str != "" {
		spl := strings.Split(str, ",")
		for _, key := range spl {
			if len(key) < 1 {
				return nil, errors.New("invalid sort parameter")
			}

			// if parameter is prefixed with '-', order is descending
			var asc bool
			if key[0] == '-' {
				asc = false
				key = key[1:]
			} else {
				asc = true
			}

			field, ok := resource.Fields[key]
			if !ok {
				return nil, errors.New(fmt.Sprintf("unknown attribute: %s", key))
			}

			// check if field has jargo:"sort" tag
			if !field.Settings.AllowSorting {
				return nil, errors.New(fmt.Sprintf("sorting by %s is disabled", key))
			}

			sorting[field] = asc
		}
	}

	return &sorting, nil
}
