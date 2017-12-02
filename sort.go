package jargo

import (
	"github.com/go-pg/pg/types"
	"github.com/go-pg/pg/orm"
	"fmt"
	"strings"
	"errors"
)

const (
	dirAscending  = "ASC"
	dirDescending = "DESC"
)

type Sorting map[types.Q]bool

func (sort *Sorting) ApplyToQuery(q *orm.Query) {
	for column, asc := range *sort {
		var dir string
		if asc {
			dir = dirAscending
		} else {
			dir = dirDescending
		}

		q.OrderExpr(fmt.Sprintf("%s %s", column, dir))
	}
}

func parseSortParameters(model *Model, str string) (*Sorting, error) {
	sorting := make(Sorting)

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

			field, ok := model.Fields[key]
			if !ok {
				return nil, errors.New(fmt.Sprintf("unknown attribute: %s", key))
			}

			// check if field has jargo:"sort" tag
			if !field.Settings.AllowSorting {
				return nil, errors.New(fmt.Sprintf("sorting by %s is disabled", key))
			}

			sorting[field.Column] = asc
		}
	}

	return &sorting, nil
}
