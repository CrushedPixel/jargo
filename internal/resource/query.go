package resource

import (
	"github.com/go-pg/pg/orm"
	"errors"
	"reflect"
	"github.com/go-pg/pg"
	"github.com/gin-gonic/gin"
)

var errQueryType = errors.New("invalid query type")
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

	return &Query{
		Query:    db.Model(model),
		typ:      typ,
		resource: resource,
		fields:   allFields(resource),
	}
}

func (q *Query) Fields(fs *FieldSet) *Query {
	if q.executed {
		panic(errAlreadyExecuted)
	}
	if fs.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.fields = fs

	return q
}

func (q *Query) Sort(s *SortFields) *Query {
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

func (q *Query) Pagination(p *Pagination) *Query {
	if q.executed {
		panic(errAlreadyExecuted)
	}
	if !q.collection {
		panic(errNoCollection)
	}
	q.pagination = p

	return q
}

func (q *Query) Filters(f *Filters) *Query {
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

// TODO: sort/page/filter for select many

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

	return q.resource.NewResponse(result, q.fields).Send(c)
}

func (q *Query) execute() {
	var err error

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
			q.fields.ApplyToQuery(q.Query)
		}
	}

	// execute query
	switch q.typ {
	case typeSelect:
		err = q.Select()
		break
	case typeInsert:
		_, err = q.Insert()
		break
	case typeUpdate:
		_, err = q.Update()
		break
	case typeDelete:
		_, err = q.Delete()
		break
	default:
		panic(errQueryType)
	}

	// create resource model and populate it with query result fields
	if q.collection {
		results := make([]interface{}, 0)
		for i := 0; i < q.value.Elem().Len(); i++ {
			val := q.value.Elem().Index(i)
			result := reflect.New(q.resource.Type)

			q.fields.applyValues(&val, &result)

			results = append(results, result)
		}

		// result is slice of struct pointers
		q.result = results
	} else {
		result := reflect.New(q.resource.Type)
		q.fields.applyValues(&q.value, &result)

		// result is struct pointer
		q.result = result
	}

	q.executed = true
	q.executionError = err
}
