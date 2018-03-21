package internal

import (
	"github.com/c9s/inflect"
	"reflect"
)

type baseField struct {
	schema *Schema

	fieldName string
	fieldType reflect.Type

	name string

	jargoWritable   bool
	jargoSortable   bool
	jargoFilterable bool

	jsonapiExported  bool
	jsonapiOmitempty bool

	sqlUnique bool

	jsonapiF []reflect.StructField
	pgF      []reflect.StructField
}

func (f *baseField) JSONAPIName() string {
	return f.name
}

func (f *baseField) Writable() bool {
	return f.jargoWritable
}

func (f *baseField) Sortable() bool {
	return f.jargoSortable && !isNullable(f.fieldType)
}

func (f *baseField) Filterable() bool {
	return f.jargoFilterable
}

func (f *baseField) jsonapiFields() []reflect.StructField {
	return f.jsonapiF
}

func (f *baseField) jsonapiJoinFields() []reflect.StructField {
	return f.jsonapiF
}

func (f *baseField) pgFields() []reflect.StructField {
	return f.pgF
}

func (f *baseField) pgJoinFields() []reflect.StructField {
	return f.pgF
}

func (f *baseField) typ() reflect.Type {
	return f.fieldType
}

func newBaseField(schema *Schema, f *reflect.StructField) *baseField {
	// determine default jsonapi member name.
	// defaults to dasherized struct field name.
	defaultName := inflect.Dasherize(f.Name)
	parsed := parseJargoTagDefaultName(f.Tag.Get(jargoFieldTag), defaultName)

	field := &baseField{
		schema:          schema,
		fieldName:       f.Name,
		fieldType:       f.Type,
		name:            parsed.Name,
		jsonapiExported: parsed.Name != unexportedFieldName,
	}

	// parse options
	field.jargoWritable = !isSet(parsed.Options, optionReadonly)
	field.jargoSortable = !isSet(parsed.Options, optionNoSort)
	field.jargoFilterable = !isSet(parsed.Options, optionNoFilter)
	field.sqlUnique = isSet(parsed.Options, optionUnique)
	field.jsonapiOmitempty = isSet(parsed.Options, optionOmitempty)
	if field.jsonapiOmitempty && !field.jsonapiExported {
		panic(errJsonapiOptionOnUnexportedField)
	}

	return field
}
