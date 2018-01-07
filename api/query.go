package api

import (
	"github.com/crushedpixel/margo"
	"github.com/go-pg/pg/orm"
)

type Query interface {
	margo.Response
	Raw() *orm.Query

	Fields(FieldSet) Query
	Sort(SortFields) Query
	Pagination(Pagination) Query
	Filters(Filters) Query

	// returns the query result resource model.
	// executes the query if it hasn't been executed yet.
	Result() (interface{}, error)

	// executes the query.
	// returns the resulting resource model.
	Execute() (interface{}, error)
}
