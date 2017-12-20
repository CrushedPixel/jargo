package api

import (
	"github.com/go-pg/pg/orm"
	"crushedpixel.net/margo"
)

type Query interface {
	margo.Response
	Raw() *orm.Query

	Fields(FieldSet) Query
	Sort(SortFields) Query
	Pagination(Pagination) Query
	Filters(Filters) Query

	GetValue() (interface{}, error)
}
