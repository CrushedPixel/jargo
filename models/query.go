package models

import (
	"github.com/go-pg/pg/orm"
	"reflect"
	"errors"
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
	Type     QueryType
	value    reflect.Value
	executed bool
}

func (q *Query) Execute() error {
	var err error

	switch q.Type {
	case Select:
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
		return ErrQueryType
	}

	q.executed = true
	return err
}

func (q *Query) GetValue() (interface{}, error) {
	if !q.executed {
		err := q.Execute()
		if err != nil {
			return nil, err
		}
	}

	return q.value.Elem().Interface(), nil
}
