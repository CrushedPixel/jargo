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
	resource   api.Resource
	collection bool // whether the resource model is a slice

	// user settable
	fields     api.FieldSet
	sort       api.SortFields
	pagination api.Pagination
	filters    api.Filters

	// internal
	executed       bool
	executionError error
	result         interface{}   // the resource model
	model          reflect.Value // reference to the pg model
}

func newQuery(db orm.DB, resource *resource, typ queryType, collection bool) *Query {
	var model interface{}
	if collection {
		val := reflect.ValueOf(resource.NewPGModelCollection())
		// get pointer to slice as expected by go-pg
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		model = ptr.Interface()
	} else {
		model = resource.NewPGModelInstance()
	}

	return newQueryWithPGModelInstance(db, resource, typ, collection, model)
}

func newQueryFromResourceModel(db orm.DB, resource *resource, typ queryType, data interface{}) *Query {
	collection := resource.IsResourceModelCollection(data)
	var pgModel interface{}
	if collection {
		instances := resource.ParseResourceModelCollection(data)
		pgInstances := make([]interface{}, len(instances))
		for i := 0; i < len(instances); i++ {
			pgInstances = append(pgInstances, instances[i].ToPGModel())
		}

		pgModel = pgInstances
	} else {
		pgModel = resource.ParseResourceModel(data).ToPGModel()
	}

	return newQueryWithPGModelInstance(db, resource, typ, collection, pgModel)
}

func newQueryWithPGModelInstance(db orm.DB, resource *resource, typ queryType, collection bool, pgModelInstance interface{}) *Query {
	return &Query{
		Query:      db.Model(pgModelInstance),
		typ:        typ,
		resource:   resource,
		collection: collection,
		fields:     allFields(resource),
		model:      reflect.ValueOf(pgModelInstance),
	}
}

func (q *Query) Raw() *orm.Query {
	return q.Query
}

func (q *Query) Fields(in api.FieldSet) api.Query {
	fs := in.(*fieldSet)
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
	s := in.(*sortFields)
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
	p := in.(*pagination)
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
	f := in.(*filters)
	if q.typ != typeSelect {
		panic(errNotSelecting)
	}
	if q.executed {
		panic(errAlreadyExecuted)
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

		if q.filters != nil {
			q.filters.ApplyToQuery(q.Query)
		}

		if q.collection {
			if q.sort != nil {
				q.sort.ApplyToQuery(q.Query)
			}
			if q.pagination != nil {
				q.pagination.ApplyToQuery(q.Query)
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

	m := q.model
	if q.collection {
		var entries []interface{}

		for i := 0; i < m.Elem().Len(); i++ {
			v := m.Elem().Index(i)
			if !v.IsNil() {
				entries = append(entries, q.resource.ParsePGModel(v.Interface()).ToResourceModel())
			}
		}
		q.result = q.resource.NewResourceModelCollection(entries...)
	} else {
		q.result = q.resource.ParsePGModel(m.Interface()).ToResourceModel()
	}
}
