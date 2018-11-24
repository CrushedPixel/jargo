package jargo

import (
	"github.com/go-pg/pg/orm"
	"strconv"
)

const keyNumber = "number"

var (
	apiErrOffsetPaginationDisabled = ErrInvalidQueryParams("offset-based pagination is disabled")
	apiErrInvalidPageNumber        = ErrInvalidQueryParams("page number must be an integer")
)

type offsetPagination struct {
	*basePagination

	// page number
	number int
}

func (p *offsetPagination) applyToQuery(q *orm.Query) *orm.Query {
	q = p.applyBase(q)
	if p.number > 0 {
		q = q.Offset(p.number * int(p.pageSize))
	}
	return q
}

// parseOffsetPagination creates an offsetPagination instance for
// the given query parameters. It returns nil if no query parameters
// for offset pagination are specified.
// It returns an error if offset pagination query parameters
// are specified and they are either invalid
// or offset pagination is disabled for the Application.
func (app *Application) parseOffsetPagination(base *basePagination, pageParams map[string]string) (Pagination, error) {
	if v, ok := pageParams[keyNumber]; ok {
		if !app.paginationStrategies.Offset {
			return nil, apiErrOffsetPaginationDisabled
		}

		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, apiErrInvalidPageNumber
		}

		return &offsetPagination{base, n}, nil
	}
	return nil, nil
}
