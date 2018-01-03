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
	"gopkg.in/go-playground/validator.v9"
)

// implements api.Resource
type resource struct {
	*schema
}

func (r *resource) ParseJsonapiPayload(in io.Reader, validate bool) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(in, r.NewResourceModelInstance(), validate)
}

func (r *resource) ParseJsonapiPayloadString(payload string, validate bool) (interface{}, error) {
	return r.ParseJsonapiPayload(strings.NewReader(payload), validate)
}

func (r *resource) ParseJsonapiUpdatePayload(in io.Reader, instance interface{}, validate bool) (interface{}, error) {
	return r.unmarshalJsonapiPayload(in, instance, validate)
}

func (r *resource) ParseJsonapiUpdatePayloadString(payload string, instance interface{}, validate bool) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(strings.NewReader(payload), instance, validate)
}

// unmarshals a jsonapi payload, applying it to a resource model instance.
// if validate is true, it also validates all fields set by the user.
func (r *resource) unmarshalJsonapiPayload(in io.Reader, resourceModelInstance interface{}, validate bool) (interface{}, error) {
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
			if validate {
				err = fieldInstance.validate()
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

func (r *resource) Validate(instance interface{}) error {
	return r.ParseResourceModel(instance).Validate()
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
