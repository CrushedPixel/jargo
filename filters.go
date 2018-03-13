package jargo

import (
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg/orm"
	"strconv"
	"strings"
)

// Filters contains filter instructions
// for a specific Resource.
// It can be used in Queries
// to filter results by certain attributes.
type Filters struct {
	resource *Resource
	filters  map[internal.SchemaField]*Filter
}

// A Filter contains values to be filtered by,
// each of the filter operators being connected
// via a logical OR, and all of the values for
// an operator being connected via a logical AND.
type Filter struct {
	Eq   []string
	Not  []string
	Like []string
	Lt   []string
	Lte  []string
	Gt   []string
	Gte  []string
}

func newFilters(r *Resource, filters map[internal.SchemaField]*Filter) *Filters {
	return &Filters{
		resource: r,
		filters:  filters,
	}
}

func (f *Filters) applyToQuery(q *orm.Query) {
	for field, filter := range f.filters {
		filter.applyToQuery(q, field)
	}
}

func (f *Filter) applyToQuery(q *orm.Query, field internal.SchemaField) {
	andWhereOr(q, field, "=", f.Eq)
	andWhereOr(q, field, "<>", f.Not)
	andWhereOr(q, field, "LIKE", f.Like)
	andWhereOr(q, field, "<", f.Lt)
	andWhereOr(q, field, "<=", f.Lte)
	andWhereOr(q, field, ">", f.Gt)
	andWhereOr(q, field, ">=", f.Gte)
}

// generates an AND WHERE (xxx OR xxx) clause
func andWhereOr(q *orm.Query, field internal.SchemaField, op string, values []string) {
	if len(values) > 0 {
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

// ParseFilters parses creates a Filters instance
// for the given map of filter parameters.
// These parameters can be created manually
// or extracted from an URL's query parameters
// using ParseFilterParameters.
//
// Returns ErrInvalidQueryParams when encountering invalid query values.
func (r *Resource) ParseFilters(parsed map[string]map[string][]string) (*Filters, error) {
	// convert parsed data into Filter map
	f := make(map[string]*Filter)
	for field, filters := range parsed {
		filter := &Filter{}
		for op, values := range filters {
			switch strings.ToUpper(op) {
			case "EQ":
				filter.Eq = append(filter.Eq, values...)
			case "NOT":
				filter.Not = append(filter.Not, values...)
			case "LIKE":
				filter.Like = append(filter.Like, values...)
			case "LT":
				filter.Lt = append(filter.Lt, values...)
			case "LTE":
				filter.Lte = append(filter.Lte, values...)
			case "GT":
				filter.Gt = append(filter.Gt, values...)
			case "GTE":
				filter.Gte = append(filter.Gte, values...)
			default:
				return nil, ErrInvalidQueryParams(fmt.Sprintf(`unknown filter operator "%s"`, op))
			}
		}
		f[field] = filter
	}

	filters, err := r.Filters(f)
	if err != nil {
		return nil, ErrInvalidQueryParams(err.Error())
	}
	return filters, nil
}

// Filters returns a Filters instance for a map of
// JSON API field names and Filter instances.
//
// Returns an error if a field is not a valid
// JSON API field name for this resource
// or a filter operator is not supported.
func (r *Resource) Filters(filters map[string]*Filter) (*Filters, error) {
	f := make(map[internal.SchemaField]*Filter)
	for fieldName, filter := range filters {
		// find resource field with matching jsonapi name
		var field internal.SchemaField
		for _, rf := range r.schema.Fields() {
			if rf.JSONAPIName() == fieldName {
				field = rf
				break
			}
		}
		if field == nil {
			return nil, fmt.Errorf(`unknown filter parameter: "%s"`, fieldName)
		}
		if !field.Filterable() {
			return nil, fmt.Errorf(`filtering by "%s" is disabled`, fieldName)
		}

		f[field] = filter
	}

	return newFilters(r, f), nil
}

// IdFilter returns a Filters instance filtering by the id field.
func (r *Resource) IdFilter(id int64) *Filters {
	f, err := r.Filters(map[string]*Filter{internal.IdFieldJsonapiName: {Eq: []string{strconv.FormatInt(id, 10)}}})
	if err != nil {
		panic(err)
	}

	return f
}

// allFields returns a FieldSet containing
// all fields for this Resource.
func (r *Resource) allFields() *FieldSet {
	return newFieldSet(r, r.schema.Fields())
}
