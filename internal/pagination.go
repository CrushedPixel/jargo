package internal

import (
	"errors"
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

type pagination struct {
	Number int // page[number]
	Size   int // page[size]
}

func (p *pagination) ApplyToQuery(q *orm.Query) {
	q.Offset(p.Number * p.Size).Limit(p.Size)
}

func ParsePagination(query url.Values, maxPageSize int) (*pagination, error) {
	parsed := parser.ParsePageParameters(query)
	return newPagination(parsed, maxPageSize)
}

func newPagination(values map[string]string, maxPageSize int) (*pagination, error) {
	p := &pagination{
		Number: 0,
		Size:   maxPageSize,
	}

	for k, v := range values {
		switch k {
		case keyNumber:
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.New(fmt.Sprintf(`value for page parameter "%s" must be an integer`, k))
			}

			p.Number = n
			break
		case keySize:
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.New(fmt.Sprintf(`value for page parameter "%s" must be an integer`, k))
			}

			if n > maxPageSize {
				return nil, errors.New(fmt.Sprintf("maximum page size is %d", maxPageSize))
			}

			p.Size = n
			break
		default:
			return nil, errors.New(fmt.Sprintf(`unknown page parameter: "%s"`, k))
		}
	}

	return p, nil
}
