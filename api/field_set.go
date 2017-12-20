package api

import "github.com/go-pg/pg/orm"

type FieldSet interface {
	ApplyToQuery(*orm.Query)
}
