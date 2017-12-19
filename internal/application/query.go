package application

import (
	"github.com/go-pg/pg/orm"
	"reflect"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"net/http"
	"crushedpixel.net/margo"
)

var ErrQueryType = errors.New("invalid query type")

type QueryType int

const (
	Select QueryType = iota + 1
	Insert
	Update
	Delete
)

type Query struct {
	*orm.Query
	Type QueryType

	// parameters used by Select queries
	Sort   ResultSorting
	Page   *ResultPagination
	Filter Filters

	Fields ResultFields // passed to DataResponse

	value          reflect.Value // reference to the model which is operated on by the orm query
	executed       bool
	executionError error
}

func newQuery(typ QueryType, db *pg.DB, instance interface{}) *Query {
	return &Query{
		Query: db.Model(instance),
		Type:  typ,
		value: reflect.ValueOf(instance),
	}
}

func newSelectQuery(db *pg.DB, instance interface{}) *Query {
	return newQuery(Select, db, instance)
}

func newInsertQuery(db *pg.DB, instance interface{}) *Query {
	return newQuery(Insert, db, instance)
}

func newUpdateQuery(db *pg.DB, instance interface{}) *Query {
	return newQuery(Update, db, instance)
}

func newDeleteQuery(db *pg.DB, instance interface{}) *Query {
	return newQuery(Delete, db, instance)
}

// implements margo.Response
func (q *Query) Send(c *gin.Context) error {
	val, err := q.GetValue()
	if err != nil {
		if err == pg.ErrNoRows {
			return ApiErrNotFound
		}
		return err
	}

	switch q.Type {
	case Select, Update:
		return NewDataResponse(val, q.Fields).Send(c)
	case Insert:
		return NewDataResponseWithStatusCode(val, q.Fields, http.StatusCreated).Send(c)
	case Delete:
		return margo.NewEmptyResponse(http.StatusNoContent).Send(c)
	}

	return ErrQueryType
}

func (q *Query) ApplyQueryParams(qp *QueryParams) *Query {
	q.Sort = qp.Sort
	q.Page = &qp.Page
	q.Fields = qp.Fields

	return q
}

func (q *Query) ApplyIndexQueryParams(qp *IndexQueryParams) *Query {
	q.Filter = qp.Filter

	return q
}

func (q *Query) execute() {
	var err error

	switch q.Type {
	case Select:
		// apply request parameters if query returns
		// a resource collection
		if q.value.Elem().Kind() == reflect.Slice {
			if q.Filter != nil {
				q.Filter.ApplyToQuery(q)
			}

			if q.Sort != nil {
				q.Sort.ApplyToQuery(q)
			}

			if q.Page != nil {
				q.Page.ApplyToQuery(q)
			}
		}

		err = q.Select()
		break
	case Insert:
		_, err = q.Insert()
		break
	case Update:
		_, err = q.Update()
		break
	case Delete:
		_, err = q.Delete()
		break
	default:
		err = ErrQueryType
	}

	q.executed = true
	q.executionError = err
}

func (q *Query) GetValue() (interface{}, error) {
	if !q.executed {
		q.execute()
	}

	if q.executionError != nil {
		return nil, q.executionError
	}

	if q.value.Elem().Kind() == reflect.Slice {
		// return slice of pointers to structs
		return q.value.Elem().Interface(), nil
	}

	// return pointer to struct
	return q.value.Interface(), nil
}
