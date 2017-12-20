package internal

import (
	"reflect"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/margo"
	"net/http"
	"github.com/gin-gonic/gin"
	"crushedpixel.net/jargo/internal/parser"
	"net/url"
)

// to be used whenever a variable that is only available
// after initialization is accessed in a function
var errNotInitialized = "resource has not been initialized yet"

type Resource struct {
	Type reflect.Type // type of the resource model

	definition  *resourceDefinition
	initialized bool

	// fields only available after initialization by Registry
	fields []*resourceField

	pgModel             reflect.Type
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
	parsed, err := parser.ParseFieldParameters(query)
	if err != nil {
		return nil, err
	}

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

func (r *Resource) CreateTable(db *pg.DB) error {
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
	q.Raw().Where("Id = ?", id)
	return q
}

func (r *Resource) ResponseWithStatusCode(data interface{}, fields api.FieldSet, status int) margo.Response {
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
	return r.Response(data, allFields(r))
}

