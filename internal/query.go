package internal

import (
	"github.com/go-pg/pg/orm"
	"errors"
	"reflect"
	"github.com/go-pg/pg"
	"github.com/gin-gonic/gin"
	"crushedpixel.net/jargo/api"
	"net/http"
	"crushedpixel.net/margo"
)

var errQueryType = errors.New("invalid query type")
var errNotSelecting = errors.New("query type must be select")
var errAlreadyExecuted = errors.New("query has already been executed")
var errNoCollection = errors.New("query must be a collection")
var errMismatchingResource = errors.New("resource does not match query resource")

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
	model          reflect.Value // reference to the pg model
}

func newQuery(db orm.DB, resource *Resource, typ queryType, collection bool) *Query {
	var model interface{}
	if collection {
		model = resource.newModelSlice()
	} else {
		model = resource.newModelInstance()
	}

	return newQueryWithPGModel(db, resource, typ, collection, model)
}

func newQueryFromResourceModel(db orm.DB, resource *Resource, typ queryType, collection bool, data interface{}) *Query {
	// convert resource model to pg model
	pgModel, err := resource.registry.resourceModelToPGModel(resource, reflect.ValueOf(data), resource.pgModel)
	if err != nil {
		panic(err)
	}

	return newQueryWithPGModel(db, resource, typ, collection, pgModel)
}

func newQueryWithPGModel(db orm.DB, resource *Resource, typ queryType, collection bool, model interface{}) *Query {
	return &Query{
		Query:      db.Model(model),
		typ:        typ,
		resource:   resource,
		collection: collection,
		fields:     allFields(resource),
		model:      reflect.ValueOf(model),
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
	if q.typ != typeSelect {
		panic(errNotSelecting)
	}
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
	if q.typ != typeSelect {
		panic(errNotSelecting)
	}
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
	if q.typ != typeSelect {
		panic(errNotSelecting)
	}
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

func (q *Query) Result() (interface{}, error) {
	if !q.executed {
		q.execute()
	}

	if q.executionError != nil {
		return nil, q.executionError
	}

	return q.result, nil
}

func (q *Query) Execute() (interface{}, error) {
	q.execute()
	if q.executionError != nil {
		return nil, q.executionError
	}

	return q.result, nil
}

// satisfies margo.Response
func (q *Query) Send(c *gin.Context) error {
	result, err := q.Result()
	if err != nil {
		if err == pg.ErrNoRows {
			return api.ErrNotFound
		}
		return err
	}

	var response margo.Response
	switch q.typ {
	case typeSelect, typeInsert, typeUpdate:
		var status int
		if q.typ == typeInsert {
			status = http.StatusCreated
		} else {
			status = http.StatusOK
		}
		response = q.resource.ResponseWithStatusCode(result, q.fields, status)
	case typeDelete:
		response = margo.NewEmptyResponse(http.StatusNoContent)
	}

	return response.Send(c)
}

func (q *Query) execute() {
	// execute query
	switch q.typ {
	case typeSelect:
		// apply query modifiers
		q.fields.ApplyToQuery(q.Query)

		if q.collection {
			if q.sort != nil {
				q.sort.ApplyToQuery(q.Query)
			}
			if q.pagination != nil {
				q.pagination.ApplyToQuery(q.Query)
			}
			if q.filters != nil {
				q.filters.ApplyToQuery(q.Query)
			}
		}

		q.executionError = q.Select()
		break
	case typeInsert:
		_, q.executionError = q.Insert()
		break
	case typeUpdate:
		_, q.executionError = q.Update()
		break
	case typeDelete:
		var result orm.Result
		result, q.executionError = q.Delete()
		if q.executionError == nil && result.RowsAffected() == 0 {
			q.executionError = pg.ErrNoRows
		}
		break
	default:
		panic(errQueryType)
	}

	q.executed = true

	if q.executionError != nil {
		return
	}

	// create resource model and populate it with query result fields
	m := q.model
	// dereference slice pointers values as expected by pgModelToResourceModel
	if q.collection {
		m = m.Elem()
	}

	q.result, q.executionError = q.resource.registry.pgModelToResourceModel(q.resource, m)
}
