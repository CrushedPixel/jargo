package jargo

import (
	"fmt"
	"github.com/crushedpixel/jargo/internal/parser"
	"github.com/go-pg/pg/orm"
	"net/url"
	"strconv"
)

const (
	keyNumber = "number"
	keySize   = "size"
)

type Pagination struct {
	Number int // page[number]
	Size   int // page[size]
}

func (p *Pagination) applyToQuery(q *orm.Query) {
	q.Offset(p.Number * p.Size).Limit(p.Size)
}

func ParsePagination(values map[string]string, maxPageSize int) (*Pagination, error) {
	p := &Pagination{
		Number: 0,
		Size:   maxPageSize,
	}

	for k, v := range values {
		switch k {
		case keyNumber:
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf(`value for page parameter "%s" must be an integer`, k)
			}

			p.Number = n
			break
		case keySize:
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf(`value for page parameter "%s" must be an integer`, k)
			}

			if n > maxPageSize {
				return nil, fmt.Errorf("maximum page size is %d", maxPageSize)
			}

			p.Size = n
			break
		default:
			return nil, fmt.Errorf(`unknown page parameter: "%s"`, k)
		}
	}

	return p, nil
}
