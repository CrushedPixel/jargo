package jargo

import (
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
	"strconv"
)

const keyCursor = "cursor"

type cursorPagination struct {
	*basePagination

	// cursor resource instance
	cursor *internal.SchemaInstance
}

// parseCursorPagination creates a cursorPagination instance for
// the given query parameters. It returns nil if no query parameters
// for cursor pagination are specified.
// It returns an error if cursor pagination query parameters
// are specified and they are either invalid
// or cursor pagination is disabled for the Application.
func (app *Application) parseCursorPagination(r *Resource, base *basePagination, pageParams map[string]string) (Pagination, error) {
	if v, ok := pageParams[keyCursor]; ok {
		if !app.paginationStrategies.Offset {
			return nil, errors.New("cursor-based pagination is disabled")
		}

		id, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return nil, errors.New("invalid page cursor")
		}

		// fetch resource instance from database
		res, err := r.SelectById(app.DB(), id).Result()
		if err != nil {
			return nil, err
		}
		if res == nil {
			return nil, errors.New("invalid page cursor")
		}
		cursor := r.schema.ParseResourceModel(res)

		return &cursorPagination{base, cursor}, nil
	}
	return nil, nil
}

func (p *cursorPagination) applyToQuery(q *orm.Query) *orm.Query {
	q = p.applyBase(q)
	q = p.keySort(q)
	return q
}

// keySort applies the keyset pagination query clauses to q.
func (p *cursorPagination) keySort(q *orm.Query) *orm.Query {
	// generate cursor WHERE clause
	q = p.appendWhere(q, 0)

	// add ORDER BY clauses
	for _, e := range p.entries {
		var dir string
		if e.asc {
			dir = "ASC"
		} else {
			dir = "DESC"
		}

		// escape column name
		column := escapePGColumn(e.field.PGFilterColumn())

		q = q.OrderExpr(fmt.Sprintf(`"%s" %s`, column, dir))
	}

	return q
}

// appendWhere appends to q the keyset pagination WHERE clause
// for the cursorPagination's order entry at index i.
//
// See https://use-the-index-luke.com/sql/partial-results/fetch-next-page
func (p *cursorPagination) appendWhere(q *orm.Query, i int) *orm.Query {
	e := p.entries[i]

	// determine sorting operator
	var op string
	if e.asc {
		op = ">"
	} else {
		op = "<"
	}

	if i+1 < len(p.entries) {
		op += "="
	}

	// escape column name
	column := escapePGColumn(e.field.PGFilterColumn())

	// get cursor value for column
	value := p.cursor.SortValue(e.field)

	q = q.Where(fmt.Sprintf(`%s %s ?`, column, op), value)
	if i+1 < len(p.entries) {
		q.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.Where(fmt.Sprintf(`"%s" != ?`, column), value)

			// append WHERE clause for next sorting instruction inside nested condition
			q = q.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				q = p.appendWhere(q, i+1)
				return q, nil
			})
			return q, nil
		})
	}

	return q
}
