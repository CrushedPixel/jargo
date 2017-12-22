package internal

import (
	"reflect"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/margo"
	"net/http"
	"crushedpixel.net/jargo/internal/parser"
	"net/url"
	"fmt"
	"io"
	"github.com/google/jsonapi"
)

// to be used whenever a variable that is only available
// after initialization is accessed in a function
var errNotInitialized = "resource has not been initialized yet"

type Resource struct {
	modelType reflect.Type // type of the resource model

	definition  *resourceDefinition
	initialized bool

	fields []*resourceField

	registry *Registry // the registry this resource is registered with

	// we need two versions of the resources pg model,
	// joinPGModel only containing fields in the database
	// (attributes and foreign ids), and pgModel containing
	// references to the relation joinPGModels.
	// this is needed to prevent circular dependencies
	// for e.g. a has/belongsTo combinations,
	// where each pgModel needs to reference the other.
	pgModel     reflect.Type // full pg model
	joinPGModel reflect.Type // pg model without relation fields

	// jsonapi model only consisting of id field
	// for reference in relations
	joinJsonapiModel reflect.Type

	staticPGFields      []reflect.StructField
	staticJsonapiFields []reflect.StructField
}

func (r *Resource) jsonapiModel(fs *FieldSet) reflect.Type {
	if !r.initialized {
		panic(errNotInitialized)
	}
	if fs.resource != r {
		panic("trying to generate jsonapi model from field set of different resource")
	}

	fields := append(r.staticJsonapiFields, fs.jsonapiFields()...)
	return reflect.StructOf(fields)
}

// returns a struct pointer
func (r *Resource) newModelInstance() interface{} {
	if !r.initialized {
		panic(errNotInitialized)
	}
	return reflect.New(r.pgModel).Interface()
}

// returns a pointer to a slice of struct pointers
func (r *Resource) newModelSlice() interface{} {
	if !r.initialized {
		panic(errNotInitialized)
	}
	return reflect.New(reflect.SliceOf(reflect.PtrTo(r.pgModel))).Interface()
}

func (r *Resource) Name() string {
	return r.definition.name
}

func (r *Resource) ParseFieldSet(query url.Values) (api.FieldSet, error) {
	parsed := parser.ParseFieldParameters(query)

	// find fields parameter with matching jsonapi name
	// and create a FieldSet from it
	for k, v := range parsed {
		if k == r.Name() {
			return newFieldSet(r, v)
		}
	}

	// if no fields parameter is specified for this resource,
	// return a FieldSet containing all fields
	return allFields(r), nil
}

func (r *Resource) ParseSortFields(query url.Values) (api.SortFields, error) {
	parsed := parser.ParseSortParameters(query)
	return newSortFields(r, parsed)
}

func (r *Resource) ParseFilters(query url.Values) (api.Filters, error) {
	parsed, err := parser.ParseFilterParameters(query)
	if err != nil {
		return nil, err
	}

	return newFilters(r, parsed)
}

func (r *Resource) ParsePayload(in io.ReadCloser) (interface{}, error) {
	fields := settableFields(r)
	modelInstance := reflect.New(r.modelType)

	return r.parsePayload(in, modelInstance, fields)
}

func (r *Resource) ParseUpdatePayload(in io.ReadCloser, instance interface{}) (interface{}, error) {
	fields := settableFields(r)
	return r.parsePayload(in, reflect.ValueOf(instance), fields)
}

// instance is a struct pointer
func (r *Resource) parsePayload(in io.ReadCloser, modelInstance reflect.Value, fields *FieldSet) (interface{}, error) {
	jsonapiInstance, err := r.registry.resourceModelToJsonapiModel(r, modelInstance, r.jsonapiModel(fields))
	if err != nil {
		return nil, err
	}

	err = jsonapi.UnmarshalPayload(in, jsonapiInstance)
	if err != nil {
		return nil, err
	}

	// convert jsonapi instance back to resource model
	return r.registry.jsonapiModelToResourceModel(r, reflect.ValueOf(jsonapiInstance))
}

func (r *Resource) CreateTable(db *pg.DB) error {
	println(fmt.Sprintf("pg model: %v", r.pgModel))
	return db.CreateTable(r.newModelInstance(), &orm.CreateTableOptions{IfNotExists: true})
}

func (r *Resource) Select(db *pg.DB) api.Query {
	return newQuery(db, r, typeSelect, true)
}

func (r *Resource) SelectOne(db *pg.DB) api.Query {
	return newQuery(db, r, typeSelect, false)
}

func (r *Resource) SelectById(db *pg.DB, id interface{}) api.Query {
	q := r.SelectOne(db)
	q.Raw().Where(fmt.Sprintf("%s = ?", primaryFieldColumn), id)
	return q
}

func (r *Resource) Insert(db *pg.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeInsert, true, instances)
}

func (r *Resource) InsertOne(db *pg.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeInsert, false, instance)
}

func (r *Resource) Update(db *pg.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeUpdate, true, instances)
}

func (r *Resource) UpdateOne(db *pg.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeUpdate, false, instance)
}

func (r *Resource) Delete(db *pg.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeDelete, true, instances)
}

func (r *Resource) DeleteOne(db *pg.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeDelete, false, instance)
}

func (r *Resource) DeleteById(db *pg.DB, id interface{}) api.Query {
	q := newQuery(db, r, typeDelete, false)
	q.Raw().Where(fmt.Sprintf("%s = ?", primaryFieldColumn), id)
	return q
}

func (r *Resource) ResponseWithStatusCode(data interface{}, fields api.FieldSet, status int) margo.Response {
	if fields == nil {
		fields = allFields(r)
	}
	return &resourceResponse{
		resource: r,
		data:     data,
		fieldSet: fields.(*FieldSet),
		status:   status,
	}
}

func (r *Resource) Response(data interface{}, fields api.FieldSet) margo.Response {
	return r.ResponseWithStatusCode(data, fields, http.StatusOK)
}

func (r *Resource) ResponseAllFields(data interface{}) margo.Response {
	return r.Response(data, nil)
}
