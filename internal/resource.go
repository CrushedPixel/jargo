package internal

import (
	"github.com/crushedpixel/jargo/api"
	"github.com/crushedpixel/margo"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// implements api.Resource
type resource struct {
	*schema
	initialized bool
}

// Initialize makes the Resource ready to use,
// creating the necessary database tables.
// If it has already been initialized,
// it is not initialized again.
func (r *resource) Initialize(db *pg.DB) error {
	if r.initialized {
		return nil
	}
	// for now, creating the table in the database
	// is all that's needed to initialize a resource
	err := r.CreateTable(db)
	if err != nil {
		return err
	}
	r.initialized = true
	return nil
}

// ParseJsonapiPayload parses a payload from a reader
// into a Resource Model Instance according to the JSON API spec.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *resource) ParseJsonapiPayload(in io.Reader, validate *validator.Validate) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(in, r.NewResourceModelInstance(), validate)
}

// ParseJsonapiPayloadString parses a payload string
// into a Resource Model Instance according to the JSON API spec.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *resource) ParseJsonapiPayloadString(payload string, validate *validator.Validate) (interface{}, error) {
	return r.ParseJsonapiPayload(strings.NewReader(payload), validate)
}

// ParseJsonapiUpdatePayload parses a payload from a reader
// into a Resource Model Instance according to the JSON API spec,
// applying it to an existing Resource Model Instance.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *resource) ParseJsonapiUpdatePayload(in io.Reader, instance interface{}, validate *validator.Validate) (interface{}, error) {
	return r.unmarshalJsonapiPayload(in, instance, validate)
}

// ParseJsonapiUpdatePayloadString parses a payload string
// into a Resource Model Instance according to the JSON API spec,
// applying it to an existing Resource Model Instance.
// If validate is not nil, it is used to validate all writable fields.
// Returns a new Resource Model Instance.
func (r *resource) ParseJsonapiUpdatePayloadString(payload string, instance interface{}, validate *validator.Validate) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(strings.NewReader(payload), instance, validate)
}

func (r *resource) unmarshalJsonapiPayload(in io.Reader, resourceModelInstance interface{}, validate *validator.Validate) (interface{}, error) {
	si := r.ParseResourceModel(resourceModelInstance).(*schemaInstance)

	// parse payload into new jsonapi instance
	jsonapiTargetInstance := r.NewJsonapiModelInstance()
	err := jsonapi.UnmarshalPayload(in, jsonapiTargetInstance)
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(jsonapiTargetInstance)
	jmi := &jsonapiModelInstance{
		schema: r.schema,
		value:  &val,
	}

	// copy original resource model fields to a new target resource model,
	// applying writable fields from parsed jsonapi model
	target := r.newResourceModelInstance()
	for _, fieldInstance := range si.fields {
		if fieldInstance.parentField().writable() {
			fieldInstance.parseJsonapiModel(jmi)

			// NOTE: this validates any writable field,
			// regardless if it has actually been set by the user
			if validate != nil {
				err = fieldInstance.validate(validate)
				if err != nil {
					if e, ok := err.(validator.ValidationErrors); ok {
						return nil, api.ErrValidationFailed(e)
					}
					return nil, err
				}
			}
		}
		fieldInstance.applyToResourceModel(target)
	}

	return target.value.Interface(), nil
}

// Validate validates a Resource Model Instance
// according to the Resource validation rules,
// using the Validate instance provided.
func (r *resource) Validate(validate *validator.Validate, instance interface{}) error {
	return r.ParseResourceModel(instance).Validate(validate)
}

// CreateTable creates the database table
// for this Resource if it doesn't exist yet.
func (r *resource) CreateTable(db *pg.DB) error {
	err := db.CreateTable(r.NewPGModelInstance(), &orm.CreateTableOptions{IfNotExists: true})
	if err != nil {
		return err
	}

	// call afterCreateTable hooks on fields
	for _, f := range r.fields {
		if afterHook, ok := f.(afterCreateTableHook); ok {
			err = afterHook.afterCreateTable(db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Select returns a new Select Many Query.
func (r *resource) Select(db orm.DB) api.Query {
	return newQuery(db, r, typeSelect, true)
}

// Select returns a new Select One Query.
func (r *resource) SelectOne(db orm.DB) api.Query {
	return newQuery(db, r, typeSelect, false)
}

// Select returns a new Select One Query
// selecting the Resource Instance with the given id.
func (r *resource) SelectById(db orm.DB, id int64) api.Query {
	q := r.SelectOne(db)
	q.Filters(idFilter(r, id))
	return q
}

// Insert returns a new Insert Many Query
// inserting the Resource Model Instances provided.
func (r *resource) Insert(db orm.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeInsert, instances)
}

// Insert returns a new Insert One Query
// inserting the Resource Model Instance provided.
func (r *resource) InsertOne(db orm.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeInsert, instance)
}

func (r *resource) updateQuery(db orm.DB, data interface{}) api.Query {
	q := newQueryFromResourceModel(db, r, typeUpdate, data)
	// some values, for example updatedAt attributes are modified on the server,
	// so the actual values have to be fetched back from the update request.
	// this should probably be optimized to only fetch values that
	// are expected to be modified by the server.
	q.Returning("*")
	return q
}

// Update returns a new Update Many Query
// updating the Resource Model Instances provided.
func (r *resource) Update(db orm.DB, instances []interface{}) api.Query {
	return r.updateQuery(db, instances)
}

// Update returns a new Update One Query
// updating the Resource Model Instance provided.
func (r *resource) UpdateOne(db orm.DB, instance interface{}) api.Query {
	return r.updateQuery(db, instance)
}

// Delete returns a new Delete Many Query
// deleting the Resource Model Instances provided.
func (r *resource) Delete(db orm.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeDelete, instances)
}

// DeleteOne returns a new Delete One Query
// deleting the Resource Model Instance provided.
func (r *resource) DeleteOne(db orm.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeDelete, instance)
}

// Delete returns a new Delete One Query
// deleting the Resource Instances with the given id.
func (r *resource) DeleteById(db orm.DB, id int64) api.Query {
	q := newQuery(db, r, typeDelete, false)
	q.Filters(idFilter(r, id))
	return q
}

// ParseFieldSet parses field query parameters according to JSON API spec,
// returning the resulting FieldSet for this Resource.
// http://jsonapi.org/format/#fetching-sparse-fieldsets
//
// Returns an error when encountering invalid query values.
func (r *resource) ParseFieldSet(query url.Values) (api.FieldSet, error) {
	return parseFieldSet(r, query)
}

// ParseSortFields parses sort query parameters according to JSON API spec,
// returning the resulting SortFields for this Resource.
// http://jsonapi.org/format/#fetching-sorting
//
// Returns an error when encountering invalid query values.
func (r *resource) ParseSortFields(query url.Values) (api.SortFields, error) {
	return parseSortFields(r, query)
}

// ParseFilters parses filter query parameters according to JSON API spec,
// returning the resulting Filters for this Resource.
// http://jsonapi.org/format/#fetching-filtering
//
// Returns an error when encountering invalid query values.
func (r *resource) ParseFilters(query url.Values) (api.Filters, error) {
	return parseFilters(r, query)
}

// Response returns a margo.Response that sends a
// Resource Model Instance according to JSON API spec,
// using the FieldSet provided.
//
// The Response returned panics when sending
// if data is not a Resource Model Instance
// or Slice of Resource Model Instances.
func (r *resource) Response(data interface{}, fieldSet api.FieldSet) margo.Response {
	return r.ResponseWithStatusCode(data, fieldSet, http.StatusOK)
}

// Response returns a margo.Response that sends a
// Resource Model Instance according to JSON API spec,
// including all model fields.
//
// The Response returned panics when sending
// if data is not a Resource Model Instance
// or Slice of Resource Model Instances.
func (r *resource) ResponseAllFields(data interface{}) margo.Response {
	return r.Response(data, nil)
}

// Response returns a margo.Response that sends a
// Resource Model Instance according to JSON API spec,
// setting the status code and using the FieldSet provided.
//
// The Response returned panics when sending
// if data is not a Resource Model Instance
// or Slice of Resource Model Instances.
func (r *resource) ResponseWithStatusCode(data interface{}, fieldSet api.FieldSet, status int) margo.Response {
	if fieldSet == nil {
		fieldSet = allFields(r)
	}
	return newResourceResponse(r, data, fieldSet, status)
}
