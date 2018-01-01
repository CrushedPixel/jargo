package internal

import (
	"io"
	"strings"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
	"crushedpixel.net/jargo/api"
	"fmt"
	"crushedpixel.net/margo"
	"net/http"
	"net/url"
)

// implements api.Resource
type resource struct {
	*schema
}

func (r *resource) ParseJsonapiPayload(in io.Reader) (interface{}, error) {
	return r.unmarshalJsonapiPayload(in, r.NewJsonapiModelInstance())
}

func (r *resource) ParseJsonapiPayloadString(payload string) (interface{}, error) {
	return r.ParseJsonapiPayload(strings.NewReader(payload))
}

func (r *resource) ParseJsonapiUpdatePayload(in io.Reader, instance interface{}) (interface{}, error) {
	return r.unmarshalJsonapiPayload(in, resourceModelToJsonapiModel(r, instance))
}

func (r *resource) ParseJsonapiUpdatePayloadString(payload string, instance interface{}) (interface{}, error) {
	return r.ParseJsonapiUpdatePayload(strings.NewReader(payload), instance)
}

func (r *resource) unmarshalJsonapiPayload(in io.Reader, jsonapiModelInstance interface{}) (interface{}, error) {
	err := jsonapi.UnmarshalPayload(in, jsonapiModelInstance)
	if err != nil {
		return nil, err
	}
	return jsonapiModelToResourceModel(r, jsonapiModelInstance), nil
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

func (r *resource) SelectById(db orm.DB, id interface{}) api.Query {
	q := r.SelectOne(db)
	q.Raw().Where(fmt.Sprintf("%s = ?", idFieldColumn), id)
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

func (r *resource) DeleteById(db orm.DB, id interface{}) api.Query {
	q := newQuery(db, r, typeDelete, false)
	q.Raw().Where(fmt.Sprintf("%s = ?", idFieldColumn), id)
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
	return r.Response(data, allFields(r))
}

func (r *resource) ResponseWithStatusCode(data interface{}, fieldSet api.FieldSet, status int) margo.Response {
	return newResourceResponse(r, data, fieldSet, status)
}
