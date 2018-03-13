package jargo

import (
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
	"strconv"
)

const keySize = "size"

// PaginationStrategies contains information about
// which pagination strategies are supported by an application.
type PaginationStrategies struct {
	// Offset determines whether
	// offset-based pagination is enabled.
	Offset bool

	// Cursor determines whether
	// cursor-based/keyset pagination is enabled
	Cursor bool
}

// Pagination is responsible for applying sorting
// and pagination settings to a Query.
type Pagination interface {
	applyToQuery(*orm.Query) *orm.Query
}

// order determines sorting order.
type order struct {
	resource *Resource

	// entries contains the order entries.
	// Their relevance is determined
	// by their position in the slice.
	entries []*orderEntry
}

type orderEntry struct {
	field internal.SchemaField
	asc   bool
}

func (s *order) applyOrder(q *orm.Query) *orm.Query {
	for _, e := range s.entries {
		var dir string
		if e.asc {
			dir = "ASC"
		} else {
			dir = "DESC"
		}

		q = q.Order(fmt.Sprintf("%s %s", e.field.PGFilterColumn(), dir))
	}
	return q
}

type pageSize int

func (p pageSize) applyPageSize(q *orm.Query) *orm.Query {
	return q.Limit(int(p))
}

type basePagination struct {
	*order
	pageSize
}

func (p *basePagination) applyBase(q *orm.Query) *orm.Query {
	q = p.applyOrder(q)
	q = p.applyPageSize(q)
	return q
}

func (app *Application) parsePageSize(pageParams map[string]string) (pageSize, error) {
	if v, ok := pageParams[keySize]; ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, errors.New("page size must be an integer")
		}

		if n > app.maxPageSize {
			return 0, fmt.Errorf("maximum page size is %d", app.maxPageSize)
		}

		return pageSize(n), nil
	}

	return pageSize(app.maxPageSize), nil
}

// parseOrder returns an order instance for a map of
// JSON API field names and sort direction (true being ascending).
// Returns an error if a field is not
// a valid JSON API field name for this resource.
func (r *Resource) parseOrder(sortParams map[string]bool) (*order, error) {
	var entries []*orderEntry

	// byId flag keeps track whether we're sorting by id
	byId := false
	for fieldName, asc := range sortParams {
		// find resource field with matching jsonapi name
		var field internal.SchemaField
		for _, rf := range r.schema.Fields() {
			if rf.JSONAPIName() == fieldName {
				if rf.JSONAPIName() == internal.IdFieldJsonapiName {
					byId = true
				}

				field = rf
				break
			}
		}
		if field == nil {
			return nil, ErrInvalidQueryParams(fmt.Sprintf(`unknown sort parameter: "%s"`, fieldName))
		}
		if !field.Sortable() {
			return nil, fmt.Errorf(`sorting by "%s" is disabled`, fieldName)
		}

		entries = append(entries, &orderEntry{field, asc})
	}

	if !byId {
		// if we're not yet sorting by id, sort by id in descending order
		// with lowest priority, so a reliable result order is guaranteed
		entries = append(entries, &orderEntry{r.schema.IdField(), false})
	}

	o := &order{
		resource: r,
		entries:  entries,
	}
	return o, nil
}

// ParsePagination creates a Pagination instance
// for the given page and sort parameters.
// These parameters can be created manually
// or extracted from an URL's query parameters
// using ParseSortParameters and ParsePageParameters.
func (r *Resource) ParsePagination(app *Application, sortParams map[string]bool, pageParams map[string]string) (Pagination, error) {
	size, err := app.parsePageSize(pageParams)
	if err != nil {
		return nil, err
	}

	order, err := r.parseOrder(sortParams)
	if err != nil {
		return nil, err
	}

	base := &basePagination{order, size}

	op, err := app.parseOffsetPagination(base, pageParams)
	if err != nil {
		return nil, err
	}
	if op != nil {
		return op, nil
	}

	cp, err := app.parseCursorPagination(r, base, pageParams)
	if err != nil {
		return nil, err
	}
	if cp != nil {
		return cp, nil
	}

	// return default pagination, which is an offset pagination
	// with no offset, effectively only limiting the page size
	return &offsetPagination{base, 0}, nil
}
