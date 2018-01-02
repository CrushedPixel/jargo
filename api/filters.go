package api

import "github.com/go-pg/pg/orm"

type Filters interface {
	ApplyToQuery(*orm.Query)
}
