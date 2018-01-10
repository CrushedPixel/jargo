package jargo

import (
	"errors"
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/crushedpixel/jargo/internal/parser"
	"github.com/crushedpixel/margo"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// implements api.Resource
type Resource struct {
	schema      *internal.Schema
	initialized bool
}

// Initialize makes the Resource ready to use,
// creating the necessary database tables.
// If it has already been initialized,
// it is not initialized again.
func (r *Resource) Initialize(db *pg.DB) error {
	if r.initialized {
		return nil
	}
	// for now, creating the table in the database
	// is all that's needed to initialize a Resource
	err := r.schema.CreateTable(db)
	if err != nil {
		return err
	}
	r.initialized = true
	return nil
}

// JSONAPIName returns the JSON API member name of the Resource.
func (r *Resource) JSONAPIName() string {
	return r.schema.JSONAPIName()
}

// ParseJsonapiPayload parses a payload from a reader
// into a Resource Model Instance according to the JSON API spec.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *Resource) ParseJsonapiPayload(in io.Reader, validate *validator.Validate) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(in, r.schema.NewResourceModelInstance(), validate)
}

// ParseJsonapiPayloadString parses a payload string
// into a Resource Model Instance according to the JSON API spec.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *Resource) ParseJsonapiPayloadString(payload string, validate *validator.Validate) (interface{}, error) {
	return r.ParseJsonapiPayload(strings.NewReader(payload), validate)
}

// ParseJsonapiUpdatePayload parses a payload from a reader
// into a Resource Model Instance according to the JSON API spec,
// applying it to an existing Resource Model Instance.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *Resource) ParseJsonapiUpdatePayload(in io.Reader, instance interface{}, validate *validator.Validate) (interface{}, error) {
	instance, err := r.schema.UnmarshalJsonapiPayload(in, instance, validate)
	if err != nil {
		if e, ok := err.(validator.ValidationErrors); ok {
			return nil, ErrValidationFailed(e)
		}
		return nil, err
	}
	return instance, nil
}

// ParseJsonapiUpdatePayloadString parses a payload string
// into a Resource Model Instance according to the JSON API spec,
// applying it to an existing Resource Model Instance.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *Resource) ParseJsonapiUpdatePayloadString(payload string, instance interface{}, validate *validator.Validate) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(strings.NewReader(payload), instance, validate)
}

// Validate validates a Resource Model Instance
// according to the Resource validation rules,
// using the Validate instance provided.
func (r *Resource) Validate(validate *validator.Validate, instance interface{}) error {
	err := r.schema.ParseResourceModel(instance).Validate(validate)
	if err != nil {
		if e, ok := err.(validator.ValidationErrors); ok {
			return ErrValidationFailed(e)
		}
	}
	return nil
}

// Select returns a new Select Many Query.
func (r *Resource) Select(db orm.DB) *Query {
	return r.newQuery(db, typeSelect, true)
}

// Select returns a new Select One Query.
func (r *Resource) SelectOne(db orm.DB) *Query {
	return r.newQuery(db, typeSelect, false)
}

// Select returns a new Select One Query
// selecting the Resource Instance with the given id.
func (r *Resource) SelectById(db orm.DB, id int64) *Query {
	q := r.SelectOne(db)
	q.Filters(r.IdFilter(id))
	return q
}

// Insert returns a new Insert Many Query
// inserting the Resource Model Instances provided.
//
// Panics if instances is not a Slice of Resource Model Instances.
func (r *Resource) Insert(db orm.DB, instances []interface{}) *Query {
	return r.newQueryFromResourceData(db, typeInsert, true, instances)
}

// Insert returns a new Insert One Query
// inserting the Resource Model Instance provided.
//
// Panics if instance is not a Resource Model Instance.
func (r *Resource) InsertOne(db orm.DB, instance interface{}) *Query {
	return r.newQueryFromResourceData(db, typeInsert, false, instance)
}

func (r *Resource) updateQuery(db orm.DB, collection bool, data interface{}) *Query {
	q := r.newQueryFromResourceData(db, typeUpdate, collection, data)
	// some values, for example updatedAt attributes are modified on the server,
	// so the actual values have to be fetched back from the update request.
	// this should probably be optimized to only fetch values that
	// are expected to be modified by the server.
	q.Returning("*")
	return q
}

// Update returns a new Update Many Query
// updating the Resource Model Instances provided.
//
// Panics if instances is not a Slice of Resource Model Instances.
func (r *Resource) Update(db orm.DB, instances []interface{}) *Query {
	return r.updateQuery(db, true, instances)
}

// Update returns a new Update One Query
// updating the Resource Model Instance provided.
//
// Panics if instance is not a Resource Model Instance.
func (r *Resource) UpdateOne(db orm.DB, instance interface{}) *Query {
	return r.updateQuery(db, false, instance)
}

// Delete returns a new Delete Many Query
// deleting the Resource Model Instances provided.
//
// Panics if instances is not a Slice of Resource Model Instances.
func (r *Resource) Delete(db orm.DB, instances []interface{}) *Query {
	return r.newQueryFromResourceData(db, typeDelete, true, instances)
}

// DeleteOne returns a new Delete One Query
// deleting the Resource Model Instance provided.
//
// Panics if instance is not a Resource Model Instance.
func (r *Resource) DeleteOne(db orm.DB, instance interface{}) *Query {
	return r.newQueryFromResourceData(db, typeDelete, false, instance)
}

// Delete returns a new Delete One Query
// deleting the Resource Instances with the given id.
func (r *Resource) DeleteById(db orm.DB, id int64) *Query {
	q := r.newQuery(db, typeDelete, false)
	q.Filters(r.IdFilter(id))
	return q
}

func (r *Resource) newQuery(db orm.DB, typ queryType, collection bool) *Query {
	var model interface{}
	if collection {
		val := reflect.ValueOf(r.schema.NewPGModelCollection())
		// get pointer to slice as expected by go-pg
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		model = ptr.Interface()
	} else {
		model = r.schema.NewPGModelInstance()
	}

	return newQuery(db, r, typ, collection, model)
}

func (r *Resource) newQueryFromResourceData(db orm.DB, typ queryType, collection bool, data interface{}) *Query {
	isCollection := r.schema.IsResourceModelCollection(data)

	var pgModel interface{}
	if collection {
		if !isCollection {
			panic(errors.New("data must be a slice of resource model instances"))
		}
		instances := r.schema.ParseResourceModelCollection(data)

		// convert resource model instances to slice of pg instances
		pgInstances := make([]interface{}, len(instances))
		for i := 0; i < len(instances); i++ {
			pgInstances = append(pgInstances, instances[i].ToPGModel())
		}
		pgModel = pgInstances
	} else {
		if isCollection {
			panic(errors.New("data must be a resource model instance"))
		}
		pgModel = r.schema.ParseResourceModel(data).ToPGModel()
	}

	return newQuery(db, r, typ, collection, pgModel)
}

// ParseFieldSet parses field query parameters according to JSON API spec,
// returning the resulting FieldSet for this Resource.
// http://jsonapi.org/format/#fetching-sparse-fieldsets
//
// Returns ErrInvalidQueryParams when encountering invalid query values.
func (r *Resource) ParseFieldSet(query url.Values) (*FieldSet, error) {
	parsed := parser.ParseFieldParameters(query)

	var schemaFields []internal.SchemaField
	// check if user specified to filter this resource's fields
	if fields, ok := parsed[r.JSONAPIName()]; ok {
		// always include the id field,
		// so it gets fetched from the database
		fields = append(fields, internal.IdFieldJsonapiName)
		for _, fieldName := range fields {
			// find resource field with matching jsonapi name
			var field internal.SchemaField
			for _, f := range r.schema.Fields() {
				if f.JSONAPIName() == fieldName {
					field = f
					break
				}
			}
			if field == nil {
				return nil, ErrInvalidQueryParams(fmt.Sprintf(`unknown field parameter: "%s"`, fieldName))
			}

			schemaFields = append(schemaFields, field)
		}
		return newFieldSet(r, schemaFields), nil
	} else {
		return r.allFields(), nil
	}
}

// allFields returns a FieldSet containing
// all fields for this Resource.
func (r *Resource) allFields() *FieldSet {
	return newFieldSet(r, r.schema.Fields())
}

// ParseSortFields parses sort query parameters according to JSON API spec,
// returning the resulting SortFields for this Resource.
// http://jsonapi.org/format/#fetching-sorting
//
// Returns ErrInvalidQueryParams when encountering invalid query values.
func (r *Resource) ParseSortFields(query url.Values) (*SortFields, error) {
	fields := parser.ParseSortParameters(query)

	sort := make(map[internal.SchemaField]bool)
	for _, fieldName := range fields {
		// skip empty sort parameters
		if len(fieldName) < 1 {
			continue
		}

		// if parameter is prefixed with '-', order is descending
		var asc bool
		if fieldName[0] == '-' {
			asc = false
			fieldName = fieldName[1:]
		} else {
			asc = true
		}

		// find resource field with matching jsonapi name
		var field internal.SchemaField
		for _, rf := range r.schema.Fields() {
			if rf.JSONAPIName() == fieldName {
				field = rf
				break
			}
		}
		if field == nil {
			return nil, ErrInvalidQueryParams(fmt.Sprintf(`unknown sort parameter: "%s"`, fieldName))
		}
		if !field.Sortable() {
			return nil, errors.New(fmt.Sprintf(`sorting by "%s" is disabled`, fieldName))
		}

		sort[field] = asc
	}

	return newSortFields(r, sort), nil
}

// ParseFilters parses filter query parameters according to JSON API spec,
// returning the resulting Filters for this Resource.
// http://jsonapi.org/format/#fetching-filtering
//
// Returns ErrInvalidQueryParams when encountering invalid query values.
func (r *Resource) ParseFilters(query url.Values) (*Filters, error) {
	parsed := parser.ParseFilterParameters(query)

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

// Filters returns a new Filters instance for a map of
// JSON API field names and Filter instances.
//
// Returns an error if an entry of fields
// is not a valid JSON API field name for this resource
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

// IdFilter returns a Filters instance filtering by the id field
func (r *Resource) IdFilter(id int64) *Filters {
	f, err := r.Filters(map[string]*Filter{internal.IdFieldJsonapiName: {Eq: []string{strconv.FormatInt(id, 10)}}})
	if err != nil {
		panic(err)
	}

	return f
}

// Response returns a margo.Response that sends a
// Resource Model Instance according to JSON API spec,
// using the FieldSet provided.
//
// Panics if data is not a Resource Model Instance
// or Slice of Resource Model Instances.
func (r *Resource) Response(data interface{}, fieldSet *FieldSet) margo.Response {
	return r.ResponseWithStatusCode(data, fieldSet, http.StatusOK)
}

// ResponseAllFields returns a margo.Response that sends a
// Resource Model Instance according to JSON API spec,
// including all model fields.
//
// Panics if data is not a Resource Model Instance
// or Slice of Resource Model Instances.
func (r *Resource) ResponseAllFields(data interface{}) margo.Response {
	return r.Response(data, nil)
}

// ResponseWithStatusCode returns a margo.Response that sends a
// Resource Model Instance according to JSON API spec,
// setting the status code and using the FieldSet provided.
//
// Panics if data is not a Resource Model Instance
// or Slice of Resource Model Instances.
func (r *Resource) ResponseWithStatusCode(data interface{}, fieldSet *FieldSet, status int) margo.Response {
	if data == nil {
		panic(errors.New("resource response data is nil"))
	}
	if fieldSet == nil {
		fieldSet = r.allFields()
	}

	// convert resource model data to jsonapi model data
	var jsonapiModelData interface{}
	collection := r.schema.IsResourceModelCollection(data)
	if collection {
		instances := r.schema.ParseResourceModelCollection(data)
		// convert each of the entries to a jsonapi model instance
		var jsonapiModels []interface{}
		for _, i := range instances {
			jsonapiModels = append(jsonapiModels, i.ToJsonapiModel())
		}
		jsonapiModelData = jsonapiModels
	} else {
		instance := r.schema.ParseResourceModel(data)
		jsonapiModelData = instance.ToJsonapiModel()
	}

	return newResourceResponse(jsonapiModelData, collection, fieldSet, status)
}
