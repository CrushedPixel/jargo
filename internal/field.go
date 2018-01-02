package internal

import "reflect"

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

	jsonapiFields() ([]reflect.StructField, error)
	jsonapiJoinFields() ([]reflect.StructField, error)

	pgFields() ([]reflect.StructField, error)
	pgJoinFields() ([]reflect.StructField, error)
}

type fieldInstance interface {
	parentField() field

	// parses a resource model instance, setting the field's value.
	parseResourceModel(*resourceModelInstance) error
	// parses a jsonapi model instance, setting the field's value.
	parseJsonapiModel(*jsonapiModelInstance) error
	// parses a resource model instance, setting the field's value.
	parsePGModel(*pgModelInstance) error

	parseJoinResourceModel(*resourceModelInstance) error
	parseJoinJsonapiModel(*joinJsonapiModelInstance) error
	parseJoinPGModel(*joinPGModelInstance) error

	// applies the field's value to a resource model instance.
	applyToResourceModel(*resourceModelInstance) error
	// applies the field's value to a resource model instance.
	applyToJsonapiModel(*jsonapiModelInstance) error
	// applies the field's value to a resource model instance.
	applyToPGModel(*pgModelInstance) error

	applyToJoinResourceModel(*resourceModelInstance) error
	applyToJoinJsonapiModel(*joinJsonapiModelInstance) error
	applyToJoinPGModel(*joinPGModelInstance) error

	validate() error
}
