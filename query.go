package jargo

import (
	"github.com/go-pg/pg/orm"
	"reflect"
	"errors"
)

var ErrQueryType = errors.New("invalid query type")

type QueryType int

const (
	Select QueryType = iota + 1
	Insert
	Update
	Delete
)

type Query struct {
	*orm.Query
	Type QueryType

	Filter Filters
	Sort   ResultSorting
	Page   *ResultPagination

	value          reflect.Value // reference to the model which is operated on by the orm query
	executed       bool
	executionError error
}

func (q *Query) ApplyQueryParams(qp *QueryParams) {
	q.Sort = qp.Sort
	q.Page = &qp.Page
}

func (q *Query) ApplyIndexQueryParams(qp *IndexQueryParams) {
	q.Filter = qp.Filter
}

func (q *Query) execute() {
	var err error

	switch q.Type {
	case Select:
		// apply request parameters if query returns
		// a resource collection
		if q.value.Elem().Kind() == reflect.Slice {
			if q.Filter != nil {
				q.Filter.ApplyToQuery(q)
			}

			if q.Sort != nil {
				q.Sort.ApplyToQuery(q)
			}

			if q.Page != nil {
				q.Page.ApplyToQuery(q)
			}
		}

		err = q.Select()
		break
	case Insert:
		_, err = q.Insert()
		break
	case Update:
		_, err = q.Update()
		break
	case Delete:
		_, err = q.Delete()
		break
	default:
		err = ErrQueryType
	}

	q.executed = true
	q.executionError = err
}

func (q *Query) GetValue() (interface{}, error) {
	if !q.executed {
		q.execute()
	}

	if q.executionError != nil {
		return nil, q.executionError
	}

	if q.value.Elem().Kind() == reflect.Slice {
		// return slice of pointers to structs
		return q.value.Elem().Interface(), nil
	}

	// return pointer to struct
	return q.value.Interface(), nil
}
