package jargo

import (
	"errors"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mohae/deepcopy"
	"net/http"
	"reflect"
)

var (
	errQueryType              = errors.New("invalid query type")
	errNotSelecting           = errors.New("query type must be select")
	errNotSelectingOrDeleting = errors.New("query type must be select or delete")
	errNoCollection           = errors.New("query must be a collection")
	errMismatchingResource    = errors.New("resource does not match query resource")
)

type queryType int

const (
	typeSelect queryType = iota + 1
	typeInsert
	typeUpdate
	typeDelete
)

// A Query is used to communicate with the database.
// It implements Response so it can be returned
// from handler functions and executed upon sending.
type Query struct {
	*orm.Query

	// final fields
	typ        queryType
	resource   *Resource
	collection bool // whether the resource model is a slice

	// user settable
	fields     *FieldSet
	pagination Pagination
	filters    *Filters

	// internal
	executed       bool
	executionError error
	result         interface{}   // the resource model
	model          reflect.Value // reference to the pg model
	response       Response      // the Response for the execution result
}

func newQuery(db orm.DB, resource *Resource, typ queryType, collection bool, pgModelInstance interface{}) *Query {
	// deepcopy the pg model instance passed
	// to ensure the original data is not being
	// modified if it has pointer references
	clone := deepcopy.Copy(pgModelInstance)

	return &Query{
		Query:      db.Model(clone),
		typ:        typ,
		resource:   resource,
		collection: collection,
		model:      reflect.ValueOf(clone),
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

// Pagination sets a Pagination instance
// to apply on Query execution.
//
// Panics if Query is not a Select many Query.
func (q *Query) Pagination(p Pagination) *Query {
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
	if q.typ != typeSelect && q.typ != typeDelete {
		panic(errNotSelectingOrDeleting)
	}
	if f.resource != q.resource {
		panic(errMismatchingResource)
	}
	q.filters = f

	return q
}

// Result returns the query result resource model.
// Executes the query if it hasn't been executed yet.
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

// Response returns the Response for the query's execution result.
// Executes the query if it hasn't been executed yet.
func (q *Query) Response() Response {
	if q.response == nil {
		result, err := q.Result()
		if err != nil {
			q.response = NewErrorResponse(err)
		} else {
			switch q.typ {
			case typeSelect:
				if !q.collection && q.result == nil {
					q.response = ErrNotFound
				} else {
					q.response = q.resource.Response(result, q.fields)
				}
			case typeInsert, typeUpdate:
				var status int
				if q.typ == typeInsert {
					status = http.StatusCreated
				} else {
					status = http.StatusOK
				}
				q.response = q.resource.ResponseWithStatusCode(result, q.fields, status)
			case typeDelete:
				if !q.collection && q.result == nil {
					q.response = ErrNotFound
				} else {
					q.response = NewResponse(http.StatusNoContent, "")
				}
			}
		}
	}

	return q.response
}

// Status satisfies Response
func (q *Query) Status() int {
	return q.Response().Status()
}

// Payload satisfies Response
func (q *Query) Payload() (string, error) {
	return q.Response().Payload()
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

		if q.collection && q.pagination != nil {
			q.pagination.applyToQuery(q.Query)
		}

		q.executionError = q.Select()
	case typeInsert:
		_, q.executionError = q.Insert()
	case typeUpdate:
		_, q.executionError = q.Update()
	case typeDelete:
		if q.filters != nil {
			q.filters.applyToQuery(q.Query)
		}

		var result orm.Result
		result, q.executionError = q.Delete()
		if q.executionError == nil && result.RowsAffected() == 0 {
			q.executionError = pg.ErrNoRows
		}
	default:
		panic(errQueryType)
	}

	q.executed = true

	if q.executionError != nil {
		// handle pg.ErrNoRows
		if q.executionError == pg.ErrNoRows {
			q.executionError = nil
			q.result = nil
			return
		}

		// handle pg.Error errors
		if pgErr, ok := q.executionError.(pg.Error); ok {
			q.executionError = q.pgErrToApiErr(pgErr)
		}

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

// pgErrToApiErr returns descriptive ApiError instances
// for specific pg.Error types. For unexpected errors,
// it returns the error itself.
//
// https://www.postgresql.org/docs/10/static/protocol-error-fields.html
// https://www.postgresql.org/docs/10/static/errcodes-appendix.html
func (q *Query) pgErrToApiErr(pgErr pg.Error) error {
	switch pgErr.Field('C') {
	case "23505": // unique_violation
		// parse constraint name to get column name
		constraint := pgErr.Field('n')
		// unique constraint is always in the format "table_column_key",
		// so we can strip away the table name and following underscore
		// from the start, and _key from the end
		column := constraint[len(q.resource.schema.Table())+1 : len(constraint)-len("_key")]

		// get json api field name for column
		var field string
		for _, f := range q.resource.schema.Fields() {
			if f.ColumnName() == column {
				field = f.JSONAPIName()
				break
			}
		}

		return newUniqueViolationError(field, column)
	default:
		return pgErr.(error)
	}
}

type UniqueViolationError struct {
	*ApiError
	// Field is the JSON API name of the field
	// whose constraint was violated
	Field string
	// Column is the database column
	// whose constraint was violated
	Column string
}

func newUniqueViolationError(field string, column string) *UniqueViolationError {
	return &UniqueViolationError{
		ApiError: NewApiError(http.StatusBadRequest, "UNIQUE_VIOLATION", field),
		Field:    field,
		Column:   column,
	}
}
