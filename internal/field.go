package internal

import (
	"reflect"
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
)

type field interface {
	createInstance() fieldInstance

	jsonapiName() string

	// pg column to use when selecting this field from the database
	pgSelectColumn() string
	// pg column to use when filtering by this field from the database
	pgFilterColumn() string

	writable() bool
	sortable() bool
	filterable() bool

	jsonapiFields() []reflect.StructField
	jsonapiJoinFields() []reflect.StructField

	pgFields() []reflect.StructField
	pgJoinFields() []reflect.StructField
}

type fieldInstance interface {
	parentField() field

	// parses a resource model instance, setting the field's value.
	parseResourceModel(*resourceModelInstance)
	// parses a jsonapi model instance, setting the field's value.
	parseJsonapiModel(*jsonapiModelInstance)
	// parses a resource model instance, setting the field's value.
	parsePGModel(*pgModelInstance)

	parseJoinResourceModel(*resourceModelInstance)
	parseJoinJsonapiModel(*joinJsonapiModelInstance)
	parseJoinPGModel(*joinPGModelInstance)

	// applies the field's value to a resource model instance.
	applyToResourceModel(*resourceModelInstance)
	// applies the field's value to a resource model instance.
	applyToJsonapiModel(*jsonapiModelInstance)
	// applies the field's value to a resource model instance.
	applyToPGModel(*pgModelInstance)

	applyToJoinResourceModel(*resourceModelInstance)
	applyToJoinJsonapiModel(*joinJsonapiModelInstance)
	applyToJoinPGModel(*joinPGModelInstance)

	validate(*validator.Validate) error
}

type afterCreateTableHook interface {
	// called on resource fields implementing afterCreateTableHook
	// by Resource.CreateTable() after table was created
	afterCreateTable(db *pg.DB) error
}
