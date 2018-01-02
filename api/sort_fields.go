package api

import "github.com/go-pg/pg/orm"

type SortFields interface {
	ApplyToQuery(*orm.Query)
}
