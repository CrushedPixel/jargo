package api

import (
	"github.com/go-pg/pg"
	"crushedpixel.net/margo"
	"net/url"
	"io"
	"github.com/go-pg/pg/orm"
)

type Resource interface {
	Name() string

	// parses a FieldSet for this Resource from fields query parameters
	ParseFieldSet(query url.Values) (FieldSet, error)
	// parses SortFields for this Resource from sort query parameters
	ParseSortFields(query url.Values) (SortFields, error)
	// parses Filters for this Resource from filter query parameters
	ParseFilters(query url.Values) (Filters, error)

	// parses a jsonapi data payload from a reader into a resource model instance
	ParsePayload(reader io.Reader) (interface{}, error)
	// parses a jsonapi data payload from a string into a resource model instance
	ParsePayloadString(payload string) (interface{}, error)
	// parses a jsonapi data payload from a reader, applying it to an existing resource model instance
	ParseUpdatePayload(reader io.Reader, instance interface{}) (interface{}, error)
	// parses a jsonapi data payload from a string, applying it to an existing resource model instance
	ParseUpdatePayloadString(payload string, instance interface{}) (interface{}, error)

	CreateTable(*pg.DB) error

	Select(orm.DB) Query
	SelectOne(orm.DB) Query
	SelectById(db orm.DB, id interface{}) Query

	Insert(db orm.DB, instances []interface{}) Query
	InsertOne(db orm.DB, instance interface{}) Query

	Update(db orm.DB, instances []interface{}) Query
	UpdateOne(db orm.DB, instance interface{}) Query

	Delete(db orm.DB, instances []interface{}) Query
	DeleteOne(db orm.DB, instance interface{}) Query
	DeleteById(db orm.DB, id interface{}) Query

	Response(data interface{}, fields FieldSet) margo.Response
	ResponseAllFields(data interface{}) margo.Response
	ResponseWithStatusCode(data interface{}, fields FieldSet, status int) margo.Response
}
