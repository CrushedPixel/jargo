package api

import (
	"github.com/crushedpixel/margo"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/url"
)

type Resource interface {
	Schema

	// parses a jsonapi data payload from a reader into a resource model instance
	// if validate is not nil, validates all writable fields
	ParseJsonapiPayload(reader io.Reader, validate *validator.Validate) (interface{}, error)
	// parses a jsonapi data payload from a string into a resource model instance
	// if validate is not nil, validates all writable fields
	ParseJsonapiPayloadString(payload string, validate *validator.Validate) (interface{}, error)
	// parses a jsonapi data payload from a reader, applying it to an existing resource model instance
	// if validate is not nil, validates all writable fields
	ParseJsonapiUpdatePayload(reader io.Reader, instance interface{}, validate *validator.Validate) (interface{}, error)
	// parses a jsonapi data payload from a string, applying it to an existing resource model instance
	// if validate is not nil, validates all writable fields
	ParseJsonapiUpdatePayloadString(payload string, instance interface{}, validate *validator.Validate) (interface{}, error)

	// validates a resource model instance according to validate struct tag,
	// using the validator.Validate provided.
	// returns the first validation errors encountered if any
	Validate(*validator.Validate, interface{}) error

	CreateTable(*pg.DB) error

	Select(orm.DB) Query
	SelectOne(orm.DB) Query
	SelectById(db orm.DB, id int64) Query

	Insert(db orm.DB, instances []interface{}) Query
	InsertOne(db orm.DB, instance interface{}) Query

	Update(db orm.DB, instances []interface{}) Query
	UpdateOne(db orm.DB, instance interface{}) Query

	Delete(db orm.DB, instances []interface{}) Query
	DeleteOne(db orm.DB, instance interface{}) Query
	DeleteById(db orm.DB, id int64) Query

	// parses a FieldSet for this Resource from fields query parameters
	ParseFieldSet(query url.Values) (FieldSet, error)
	// parses SortFields for this Resource from sort query parameters
	ParseSortFields(query url.Values) (SortFields, error)
	// parses Filters for this Resource from filter query parameters
	ParseFilters(query url.Values) (Filters, error)

	Response(interface{}, FieldSet) margo.Response
	ResponseAllFields(interface{}) margo.Response
	ResponseWithStatusCode(interface{}, FieldSet, int) margo.Response
}

type Schema interface {
	// jsonapi type name
	Name() string

	IsResourceModelCollection(interface{}) bool
	IsJsonapiModelCollection(interface{}) bool
	IsPGModelCollection(interface{}) bool

	// returns a pointer to a new resource model instance
	NewResourceModelInstance() interface{}
	// returns a slice of pointers to resource model instances
	NewResourceModelCollection(entries ...interface{}) interface{}
	// returns a pointer to a new jsonapi model instance
	NewJsonapiModelInstance() interface{}
	// returns a pointer to a new pg model instance
	NewPGModelInstance() interface{}
	// returns a slice of pointers to pg model instances
	NewPGModelCollection(entries ...interface{}) interface{}

	ParseResourceModelCollection(interface{}) []SchemaInstance
	ParseJsonapiModelCollection(interface{}) []SchemaInstance
	ParsePGModelCollection(interface{}) []SchemaInstance

	// creates a new schema instance from a resource model instance
	ParseResourceModel(interface{}) SchemaInstance
	// creates a new schema instance from a jsonapi model instance
	ParseJsonapiModel(interface{}) SchemaInstance
	// creates a new schema instance from a pg model instance
	ParsePGModel(interface{}) SchemaInstance

	ParseJoinResourceModel(interface{}) SchemaInstance
	ParseJoinJsonapiModel(interface{}) SchemaInstance
	ParseJoinPGModel(interface{}) SchemaInstance
}

type SchemaInstance interface {
	// creates a new resource model instance from this schema instance.
	ToResourceModel() interface{}
	// creates a new jsonapi model instance from this schema instance.
	ToJsonapiModel() interface{}
	// creates a new pg model instance from this schema instance.
	ToPGModel() interface{}

	ToJoinResourceModel() interface{}
	ToJoinJsonapiModel() interface{}
	ToJoinPGModel() interface{}

	// validates a schema instance according to validator rules
	Validate(*validator.Validate) error
}
