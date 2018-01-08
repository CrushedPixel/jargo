package api

import (
	"github.com/crushedpixel/margo"
	"github.com/go-pg/pg/orm"
)

// A Query is used to communicate with the database.
// It implements margo.Response so it can be returned
// from handler functions and executed upon sending.
type Query interface {
	margo.Response

	// Raw returns the wrapped orm.Query
	Raw() *orm.Query

	// Fields sets a FieldSet instance
	// to apply on Query execution.
	// FieldSets are also applied to JSON API
	// payloads created in the Send method.
	Fields(FieldSet) Query

	// Sort sets a SortFields instance
	// to apply on Query execution.
	//
	// Panics if Query is not a Select many Query.
	Sort(SortFields) Query
	// Pagination sets a Pagination instance
	// to apply on Query execution.
	//
	// Panics if Query is not a Select many Query.
	Pagination(Pagination) Query
	// Filters sets a Filters instance
	// to apply on Query execution.
	//
	// Panics if Query is not a Select Query.
	Filters(Filters) Query

	// Result returns the query result resource model.
	// It executes the query if it hasn't been executed yet.
	Result() (interface{}, error)

	// Execute executes the query and
	// returns the resulting resource model instance.
	Execute() (interface{}, error)
}
