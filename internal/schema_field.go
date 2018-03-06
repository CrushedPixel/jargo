package internal

import (
	"github.com/go-pg/pg"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

type SchemaField interface {
	createInstance() schemaFieldInstance

	// JSONAPIName returns the field's JSON API member name
	JSONAPIName() string

	// PGSelectColumn returns the pg column to use
	// when selecting this field from the database
	PGSelectColumn() string
	// PGFilterColumn returns the pg column to use
	// when filtering by this field from the database
	PGFilterColumn() string

	// Writable returns whether API users may
	// change the value of this field.
	Writable() bool
	// Sortable returns whether API users may
	// sort by this field.
	Sortable() bool
	// Filterable returns whether API users may
	// filter by this field.
	Filterable() bool

	jsonapiFields() []reflect.StructField
	jsonapiJoinFields() []reflect.StructField

	pgFields() []reflect.StructField
	pgJoinFields() []reflect.StructField
}

type schemaFieldInstance interface {
	parentField() SchemaField

	// sortValue returns the field instance's value for use when sorting.
	// Only implemented for fields that may be sortable.
	sortValue() interface{}

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
