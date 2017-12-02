package jargo

import (
	"github.com/go-pg/pg/orm"
	"errors"
	"fmt"
	"strconv"
)

type strategy int

// TODO: support for cursor-based pagination?

const (
	page   strategy = iota + 1
	cursor
)

const (
	// page strategy
	keyNumber = "number"
	keySize   = "size"
)

type Pagination struct {
	strategy strategy

	Number int // page[number]
	Size   int // page[size]
}

func (p *Pagination) ApplyToQuery(q *orm.Query) {
	switch p.strategy {
	case page:
		q.Offset(p.Number * p.Size).Limit(p.Size)
		break
	}
}

func parsePageParameters(application *Application, values map[string]string) (*Pagination, error) {
	pagination := &Pagination{}

	pagination.Size = application.MaxPageSize

	for k, v := range values {
		switch k {
		case keyNumber:
			if err := pagination.assignStrategy(page); err != nil {
				return nil, err
			}

			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("value for page parameter %s must be an integer", k))
			}

			pagination.Number = n
			break
		case keySize:
			if err := pagination.assignStrategy(page); err != nil {
				return nil, err
			}

			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("value for page parameter %s must be an integer", k))
			}

			if n > application.MaxPageSize {
				return nil, errors.New(fmt.Sprintf("maximum page size is %d", application.MaxPageSize))
			}

			pagination.Size = n
			break
		default:
			return nil, errors.New(fmt.Sprintf("unknown page parameter: %s", k))
		}
	}

	return pagination, nil
}

func (p *Pagination) assignStrategy(strategy strategy) error {
	if p.strategy != 0 && p.strategy != strategy {
		return errors.New("pagination strategies can't be mixed")
	}
	p.strategy = strategy
	return nil
}
