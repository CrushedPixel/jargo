package api

import (
	"github.com/go-pg/pg"
	"crushedpixel.net/margo"
	"net/url"
	"io"
	"github.com/go-pg/pg/orm"
)

type Resource interface {
	Schema

	// parses a jsonapi data payload from a reader into a resource model instance
	ParseJsonapiPayload(reader io.Reader, validate bool) (interface{}, error)
	// parses a jsonapi data payload from a string into a resource model instance
	ParseJsonapiPayloadString(payload string, validate bool) (interface{}, error)
	// parses a jsonapi data payload from a reader, applying it to an existing resource model instance
	ParseJsonapiUpdatePayload(reader io.Reader, instance interface{}, validate bool) (interface{}, error)
	// parses a jsonapi data payload from a string, applying it to an existing resource model instance
	ParseJsonapiUpdatePayloadString(payload string, instance interface{}, validate bool) (interface{}, error)

	// validates a resource model instance according to validate struct tag,
	// returning the first validation error encountered if any
	Validate(interface{}) error

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

	IsResourceModelCollection(interface{}) (bool, error)
	IsJsonapiModelCollection(interface{}) (bool, error)
	IsPGModelCollection(interface{}) (bool, error)

	// returns a pointer to a new resource model instance
	NewResourceModelInstance() interface{}
	// returns a slice of pointers to resource model instances
	NewResourceModelCollection(entries... interface{}) interface{}
	// returns a pointer to a new jsonapi model instance
	NewJsonapiModelInstance() interface{}
	// returns a pointer to a new pg model instance
	NewPGModelInstance() interface{}
	// returns a slice of pointers to pg model instances
	NewPGModelCollection(entries... interface{}) interface{}

	ParseResourceModelCollection(interface{}) ([]SchemaInstance, error)
	ParseJsonapiModelCollection(interface{}) ([]SchemaInstance, error)
	ParsePGModelCollection(interface{}) ([]SchemaInstance, error)

	// creates a new schema instance from a resource model instance
	ParseResourceModel(interface{}) (SchemaInstance, error)
	// creates a new schema instance from a jsonapi model instance
	ParseJsonapiModel(interface{}) (SchemaInstance, error)
	// creates a new schema instance from a pg model instance
	ParsePGModel(interface{}) (SchemaInstance, error)

	ParseJoinResourceModel(interface{}) (SchemaInstance, error)
	ParseJoinJsonapiModel(interface{}) (SchemaInstance, error)
	ParseJoinPGModel(interface{}) (SchemaInstance, error)
}

type SchemaInstance interface {
	// creates a new resource model instance from this schema instance.
	ToResourceModel() (interface{}, error)
	// creates a new jsonapi model instance from this schema instance.
	ToJsonapiModel() (interface{}, error)
	// creates a new pg model instance from this schema instance.
	ToPGModel() (interface{}, error)

	ToJoinResourceModel() (interface{}, error)
	ToJoinJsonapiModel() (interface{}, error)
	ToJoinPGModel() (interface{}, error)

	// validates a schema instance according to validator rules
	Validate() error
}
