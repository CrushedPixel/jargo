package internal

import (
	"io"
	"strings"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
	"crushedpixel.net/jargo/api"
	"crushedpixel.net/margo"
	"net/http"
	"net/url"
	"reflect"
)

// implements api.Resource
type resource struct {
	*schema
}

func (r *resource) ParseJsonapiPayload(in io.Reader) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(in, r.NewResourceModelInstance())
}

func (r *resource) ParseJsonapiPayloadString(payload string) (interface{}, error) {
	return r.ParseJsonapiPayload(strings.NewReader(payload))
}

func (r *resource) ParseJsonapiUpdatePayload(in io.Reader, instance interface{}) (interface{}, error) {
	return r.unmarshalJsonapiPayload(in, instance)
}

func (r *resource) ParseJsonapiUpdatePayloadString(payload string, instance interface{}) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(strings.NewReader(payload), instance)
}

// unmarshals a jsonapi payload, applying it to a resource model instance
func (r *resource) unmarshalJsonapiPayload(in io.Reader, resourceModelInstance interface{}) (interface{}, error) {
	si, err := r.ParseResourceModel(resourceModelInstance)
	if err != nil {
		return nil, err
	}

	// parse payload into new jsonapi instance
	jsonapiTargetInstance := r.NewJsonapiModelInstance()
	err = jsonapi.UnmarshalPayload(in, jsonapiTargetInstance)
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
	for _, fieldInstance := range si.(*schemaInstance).fields {
		if fieldInstance.parentField().writable() {
			err := fieldInstance.parseJsonapiModel(jmi)
			if err != nil {
				return nil, err
			}
		}
		fieldInstance.applyToResourceModel(target)
	}

	return target.value.Interface(), nil
}

func (r *resource) CreateTable(db *pg.DB) error {
	return db.CreateTable(r.NewPGModelInstance(), &orm.CreateTableOptions{IfNotExists: true})
}

func (r *resource) Select(db orm.DB) api.Query {
	return newQuery(db, r, typeSelect, true)
}

func (r *resource) SelectOne(db orm.DB) api.Query {
	return newQuery(db, r, typeSelect, false)
}

func (r *resource) SelectById(db orm.DB, id int64) api.Query {
	q := r.SelectOne(db)
	q.Filters(idFilter(r, id))
	return q
}

func (r *resource) Insert(db orm.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeInsert, instances)
}

func (r *resource) InsertOne(db orm.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeInsert, instance)
}

func (r *resource) Update(db orm.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeUpdate, instances)
}

func (r *resource) UpdateOne(db orm.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeUpdate, instance)
}

func (r *resource) Delete(db orm.DB, instances []interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeDelete, instances)
}

func (r *resource) DeleteOne(db orm.DB, instance interface{}) api.Query {
	return newQueryFromResourceModel(db, r, typeDelete, instance)
}

func (r *resource) DeleteById(db orm.DB, id int64) api.Query {
	q := newQuery(db, r, typeDelete, false)
	q.Filters(idFilter(r, id))
	return q
}

func (r *resource) ParseFieldSet(query url.Values) (api.FieldSet, error) {
	return parseFieldSet(r, query)
}

func (r *resource) ParseSortFields(query url.Values) (api.SortFields, error) {
	return parseSortFields(r, query)
}

func (r *resource) ParseFilters(query url.Values) (api.Filters, error) {
	return parseFilters(r, query)
}

func (r *resource) Response(data interface{}, fieldSet api.FieldSet) margo.Response {
	return r.ResponseWithStatusCode(data, fieldSet, http.StatusOK)
}

func (r *resource) ResponseAllFields(data interface{}) margo.Response {
	return r.Response(data, nil)
}

func (r *resource) ResponseWithStatusCode(data interface{}, fieldSet api.FieldSet, status int) margo.Response {
	if fieldSet == nil {
		fieldSet = allFields(r)
	}
	return newResourceResponse(r, data, fieldSet, status)
}
