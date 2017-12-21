package internal

import (
	"github.com/go-pg/pg/orm"
	"errors"
	"reflect"
	"github.com/go-pg/pg"
	"github.com/gin-gonic/gin"
	"crushedpixel.net/jargo/api"
	"net/http"
	"fmt"
)

var errQueryType = errors.New("invalid query type")
var errAlreadyExecuted = errors.New("query has already been executed")
var errNoCollection = errors.New("query must be a collection")
var errMismatchingResource = errors.New("resource does not match query resource")
var errMismatchingModelType = errors.New("model type does not match resource model type")

type queryType int

const (
	typeSelect queryType = iota + 1
	typeInsert
	typeUpdate
	typeDelete
)

// implements api.Query
type Query struct {
	*orm.Query

	// final fields
	typ        queryType
	resource   *Resource
	collection bool // whether the resource model is a slice

	// user settable
	fields     *FieldSet
	sort       *SortFields
	pagination *Pagination
	filters    *Filters

	// internal
	executed       bool
	executionError error
	result         interface{}   // the resource model
	value          reflect.Value // reference to the orm.Query model
}

func newQuery(db *pg.DB, resource *Resource, typ queryType, collection bool) *Query {
	var model interface{}
	if collection {
		model = resource.newModelSlice()
	} else {
		model = resource.newModelInstance()
	}

	return newQueryWithModel(db, resource, typ, collection, model)
}

func newQueryWithData(db *pg.DB, resource *Resource, typ queryType, collection bool, data interface{}) *Query {
	var model interface{}
	if collection {
		if reflect.TypeOf(data) != reflect.SliceOf(reflect.PtrTo(resource.modelType)) {
			panic(errMismatchingModelType)
		}
		model = resource.newModelSlice()
	} else {
		if reflect.TypeOf(data) != reflect.PtrTo(resource.modelType) {
			panic(errMismatchingModelType)
		}
		model = resource.newModelInstance()
	}

	// TODO: check if this makes sense
	// create pg model instance and apply data values to it
	modelValue := reflect.ValueOf(model)
	dataValue := reflect.ValueOf(data)

	allFields(resource).applyValues(&dataValue, &modelValue)
	return newQueryWithModel(db, resource, typ, collection, modelValue.Interface())
}

func newQueryWithModel(db *pg.DB, resource *Resource, typ queryType, collection bool, model interface{}) *Query {
	println(fmt.Sprintf("model %v", model))
	val := reflect.ValueOf(model)
	println(fmt.Sprintf("value %v", val))
	return &Query{
		Query:      db.Model(model),
		typ:        typ,
		resource:   resource,
		collection: collection,
		fields:     allFields(resource),
		value:      val,
	}
}

func (q *Query) Raw() *orm.Query {
	return q.Query
}

func (q *Query) Fields(in api.FieldSet) api.Query {
	fs := in.(*FieldSet)
	if q.executed {
		panic(errAlreadyExecuted)
	}
	if fs.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.fields = fs

	return q
}

func (q *Query) Sort(in api.SortFields) api.Query {
	s := in.(*SortFields)
	if q.executed {
		panic(errAlreadyExecuted)
	}
	if !q.collection {
		panic(errNoCollection)
	}
	if s.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.sort = s

	return q
}

func (q *Query) Pagination(in api.Pagination) api.Query {
	p := in.(*Pagination)
	if q.executed {
		panic(errAlreadyExecuted)
	}
	if !q.collection {
		panic(errNoCollection)
	}
	q.pagination = p

	return q
}

func (q *Query) Filters(in api.Filters) api.Query {
	f := in.(*Filters)
	if q.executed {
		panic(errAlreadyExecuted)
	}
	if !q.collection {
		panic(errNoCollection)
	}
	if f.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.filters = f

	return q
}

func (q *Query) GetValue() (interface{}, error) {
	if !q.executed {
		q.execute()
	}

	if q.executionError != nil {
		return nil, q.executionError
	}

	return q.result, nil
}

// satisfies margo.Response
func (q *Query) Send(c *gin.Context) error {
	result, err := q.GetValue()
	if err != nil {
		return err
	}

	var status int
	if q.typ == typeInsert {
		status = http.StatusCreated
	} else {
		status = http.StatusOK
	}

	return q.resource.ResponseWithStatusCode(result, q.fields, status).Send(c)
}

func (q *Query) execute() {
	// apply query modifiers
	q.fields.ApplyToQuery(q.Query)

	if q.collection {
		if q.sort != nil {
			q.sort.ApplyToQuery(q.Query)
		}
		if q.pagination != nil {
			q.pagination.ApplyToQuery(q.Query)
		}
		if q.fields != nil {
			q.filters.ApplyToQuery(q.Query)
		}
	}

	// execute query
	switch q.typ {
	case typeSelect:
		q.executionError = q.Select()
		break
	case typeInsert:
		_, q.executionError = q.Insert()
		break
	case typeUpdate:
		_, q.executionError = q.Update()
		break
	case typeDelete:
		_, q.executionError = q.Delete()
		break
	default:
		panic(errQueryType)
	}

	q.executed = true

	if q.executionError != nil {
		return
	}

	// create resource model and populate it with query result fields
	val := q.value
	// dereference slice pointers values as expected by pgModelToResourceModel
	if q.collection {
		val = val.Elem()
	}

	q.result, q.executionError = q.resource.registry.pgModelToResourceModel(q.resource, val)
}
