package api

import (
	"github.com/go-pg/pg/orm"
)

type Pagination interface {
	// ApplyToQuery applies the Pagination
	// parameters to an orm.Query.
	ApplyToQuery(*orm.Query)
}
