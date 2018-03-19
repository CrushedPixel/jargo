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

	// ColumnName returns the field's database column name.
	// For query building, use PGSelectColumn() and PGFilterColumn().
	ColumnName() string

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

	// typ returns the field's data type
	typ() reflect.Type
}

type schemaFieldInstance interface {
	parentField() SchemaField

	// sortValue returns the field instance's value for use when sorting.
	// Only implemented for fields that may be sortable.
	sortValue() interface{}

	// parses a resource model instance, storing it's value for the schema field.
	parseResourceModel(*resourceModelInstance)
	// parses a jsonapi model instance, storing it's value for the schema field.
	parseJsonapiModel(*jsonapiModelInstance)
	// parses a resource model instance, storing it's value for the schema field.
	parsePGModel(*pgModelInstance)

	parseJoinResourceModel(*resourceModelInstance)
	parseJoinJsonapiModel(*joinJsonapiModelInstance)
	parseJoinPGModel(*joinPGModelInstance)

	// applies the stored value to a resource model instance.
	applyToResourceModel(*resourceModelInstance)
	// applies the stored value to a jsonapi model instance.
	applyToJsonapiModel(*jsonapiModelInstance)
	// applies the stored value to a pg model instance.
	applyToPGModel(*pgModelInstance)

	applyToJoinResourceModel(*resourceModelInstance)
	applyToJoinJsonapiModel(*joinJsonapiModelInstance)
	applyToJoinPGModel(*joinPGModelInstance)

	validate(*validator.Validate) error
}

type beforeCreateTableHook interface {
	// called on resource fields implementing beforeCreateTableHook
	// by Resource.CreateTable() before table is created
	beforeCreateTable(db *pg.DB) error
}

type afterCreateTableHook interface {
	// called on resource fields implementing afterCreateTableHook
	// by Resource.CreateTable() after table was created
	afterCreateTable(db *pg.DB) error
}
