package jargo

import (
	"errors"
	"github.com/crushedpixel/margo"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"net/http"
	"reflect"
)

var errQueryType = errors.New("invalid query type")
var errNotSelecting = errors.New("query type must be select")
var errNoCollection = errors.New("query must be a collection")
var errMismatchingResource = errors.New("resource does not match query resource")

type queryType int

const (
	typeSelect queryType = iota + 1
	typeInsert
	typeUpdate
	typeDelete
)

// A Query is used to communicate with the database.
// It implements margo.Response so it can be returned
// from handler functions and executed upon sending.
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

func newQuery(db orm.DB, resource *Resource, typ queryType, collection bool, pgModelInstance interface{}) *Query {
	return &Query{
		Query:      db.Model(pgModelInstance),
		typ:        typ,
		resource:   resource,
		collection: collection,
		model:      reflect.ValueOf(pgModelInstance),
	}
}

// Fields sets a FieldSet instance
// to apply on Query execution.
// FieldSets are also applied to JSON API
// payloads created in the Send method.
func (q *Query) Fields(fs *FieldSet) *Query {
	if fs.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.fields = fs

	return q
}

// Sort sets a SortFields instance
// to apply on Query execution.
//
// Panics if Query is not a Select many Query.
func (q *Query) Sort(s *SortFields) *Query {
	if q.typ != typeSelect {
		panic(errNotSelecting)
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

// Pagination sets a Pagination instance
// to apply on Query execution.
//
// Panics if Query is not a Select many Query.
func (q *Query) Pagination(p *Pagination) *Query {
	if q.typ != typeSelect {
		panic(errNotSelecting)
	}
	if !q.collection {
		panic(errNoCollection)
	}
	q.pagination = p

	return q
}

// Filters sets a Filters instance
// to apply on Query execution.
//
// Panics if Query is not a Select Query.
func (q *Query) Filters(f *Filters) *Query {
	if q.typ != typeSelect {
		panic(errNotSelecting)
	}
	if f.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.filters = f

	return q
}

// Result returns the query result resource model.
// It executes the query if it hasn't been executed yet.
func (q *Query) Result() (interface{}, error) {
	if !q.executed {
		q.execute()
	}

	if q.executionError != nil {
		return nil, q.executionError
	}

	return q.result, nil
}

// Execute executes the query and
// returns the resulting resource model instance.
func (q *Query) Execute() (interface{}, error) {
	q.execute()
	if q.executionError != nil {
		return nil, q.executionError
	}

	return q.result, nil
}

// Send satisfies margo.Response
func (q *Query) Send(c *gin.Context) error {
	result, err := q.Result()
	if err != nil {
		if err == pg.ErrNoRows {
			return ErrNotFound
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
		response = margo.Empty(http.StatusNoContent)
	}

	return response.Send(c)
}

func (q *Query) execute() {
	// execute query
	switch q.typ {
	case typeSelect:
		// apply query modifiers
		var fields *FieldSet
		if q.fields != nil {
			fields = q.fields
		} else {
			fields = q.resource.allFields()
		}
		fields.applyToQuery(q.Query)

		if q.filters != nil {
			q.filters.applyToQuery(q.Query)
		}

		if q.collection {
			if q.sort != nil {
				q.sort.applyToQuery(q.Query)
			}
			if q.pagination != nil {
				q.pagination.applyToQuery(q.Query)
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
				entries = append(entries, q.resource.schema.ParsePGModel(v.Interface()).ToResourceModel())
			}
		}
		q.result = q.resource.schema.NewResourceModelCollection(entries...)
	} else {
		q.result = q.resource.schema.ParsePGModel(m.Interface()).ToResourceModel()
	}
}
