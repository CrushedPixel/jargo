package api

import "github.com/go-pg/pg/orm"

type Pagination interface {
	ApplyToQuery(*orm.Query)
}
