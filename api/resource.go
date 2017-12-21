package api

import (
	"github.com/go-pg/pg"
	"crushedpixel.net/margo"
	"net/url"
	"io"
)

type Resource interface {
	Name() string

	// parses a FieldSet for this Resource from fields query parameters
	ParseFieldSet(query url.Values) (FieldSet, error)
	// parses SortFields for this Resource from sort query parameters
	ParseSortFields(query url.Values) (SortFields, error)
	// parses Filters for this Resource from filter query parameters
	ParseFilters(query url.Values) (Filters, error)
	// parses a jsonapi data payload
	ParsePayload(io.ReadCloser) (interface{}, error)
	// parses a jsonapi data payload, applying it to an existing instance
	ParseUpdatePayload(in io.ReadCloser, instance interface{}) error

	CreateTable(*pg.DB) error

	Select(*pg.DB) Query
	SelectOne(*pg.DB) Query
	SelectById(db *pg.DB, id interface{}) Query

	Insert(db *pg.DB, instances []interface{}) Query
	InsertOne(db *pg.DB, instance interface{}) Query

	Update(db *pg.DB, instances []interface{}) Query
	UpdateOne(db *pg.DB, instance interface{}) Query

	Delete(db *pg.DB, instances []interface{}) Query
	DeleteOne(db *pg.DB, instance interface{}) Query
	DeleteById(db *pg.DB, id interface{}) Query

	Response(data interface{}, fields FieldSet) margo.Response
	ResponseAllFields(data interface{}) margo.Response
	ResponseWithStatusCode(data interface{}, fields FieldSet, status int) margo.Response
}
