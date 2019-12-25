package jargo

import (
	"github.com/go-pg/pg"
	"github.com/go-playground/validator/v10"
)

// DefaultMaxPageSize is the default maximum number
// of allowed entries per page.
const DefaultMaxPageSize = 25

// Options is used to configure an Application when creating it.
type Options struct {
	// DB is required.
	DB *pg.DB

	PaginationStrategies *PaginationStrategies
	MaxPageSize          int

	Validate *validator.Validate
}

func (o *Options) setDefaults() {
	if o.DB == nil {
		panic("database handle must be non-nil")
	}

	if o.PaginationStrategies == nil {
		o.PaginationStrategies = &PaginationStrategies{
			Offset: false,
			Cursor: true,
		}
	}

	if o.MaxPageSize == 0 {
		o.MaxPageSize = DefaultMaxPageSize
	}
	if o.MaxPageSize < 1 {
		panic("maximum page size has to be positive")
	}

	if o.Validate == nil {
		o.Validate = validator.New()
	}
}
