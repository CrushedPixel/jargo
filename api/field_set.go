package api

import (
	"github.com/go-pg/pg/orm"
	"github.com/google/jsonapi"
)

type FieldSet interface {
	ApplyToQuery(*orm.Query)
	ApplyToJsonapiNode(*jsonapi.Node)
}
